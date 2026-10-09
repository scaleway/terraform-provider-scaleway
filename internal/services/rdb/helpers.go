package rdb

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"strings"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

const (
	defaultInstanceTimeout       = 30 * time.Minute
	defaultInstanceUpdateTimeout = 60 * time.Minute
	defaultWaitRetryInterval     = 30 * time.Second
)

// newAPI returns a new RDB API
func newAPI(m any) *rdb.API {
	return rdb.NewAPI(meta.ExtractScwClient(m))
}

// newAPIWithRegion returns a new lb API and the region for a Create request
func newAPIWithRegion(d *schema.ResourceData, m any) (*rdb.API, scw.Region, error) {
	region, err := meta.ExtractRegion(d, m)
	if err != nil {
		return nil, "", err
	}

	return newAPI(m), region, nil
}

// NewAPIWithRegionAndID returns an lb API with region and ID extracted from the state
func NewAPIWithRegionAndID(m any, id string) (*rdb.API, scw.Region, string, error) {
	region, ID, err := regional.ParseID(id)
	if err != nil {
		return nil, "", "", err
	}

	return newAPI(m), region, ID, nil
}

// PrivilegeV1SchemaUpgradeFunc allow upgrade the privilege ID on schema V1
func PrivilegeV1SchemaUpgradeFunc(_ context.Context, rawState map[string]any, m any) (map[string]any, error) {
	idRaw, exist := rawState["id"]
	if !exist {
		return nil, errors.New("upgrade: id not exist")
	}

	idParts := strings.Split(idRaw.(string), "/")
	if len(idParts) == 4 {
		return rawState, nil
	}

	region, idStr, err := regional.ParseID(idRaw.(string))
	if err != nil {
		// force the default region
		defaultRegion, exist := meta.ExtractScwClient(m).GetDefaultRegion()
		if exist {
			region = defaultRegion
		}
	}

	databaseName := rawState["database_name"].(string)
	userName := rawState["user_name"].(string)
	rawState["id"] = ResourceRdbUserPrivilegeID(region, idStr, databaseName, userName)
	rawState["region"] = region.String()

	return rawState, nil
}

func rdbPrivilegeUpgradeV1SchemaType() cty.Type {
	return cty.Object(map[string]cty.Type{
		"id": cty.String,
	})
}

func getIPConfigCreate(d *schema.ResourceData, ipFieldName string) (ipamConfig *bool, staticConfig *string) {
	enableIpam, enableIpamSet := d.GetOk("private_network.0.enable_ipam")
	if enableIpamSet {
		ipamConfig = types.ExpandBoolPtr(enableIpam)
	}

	customIP, customIPSet := d.GetOk("private_network.0." + ipFieldName)
	if customIPSet {
		staticConfig = types.ExpandStringPtr(customIP)
	}

	return ipamConfig, staticConfig
}

// getIPConfigUpdate forces the provider to read the user's config instead of checking the state, because "enable_ipam" is not readable from the API
func getIPConfigUpdate(d *schema.ResourceData, ipFieldName string) (ipamConfig *bool, staticConfig *string) {
	if ipamConfigI, _ := meta.GetRawConfigForKey(d, "private_network.#.enable_ipam", cty.Bool); ipamConfigI != nil {
		ipamConfig = types.ExpandBoolPtr(ipamConfigI)
	}

	if staticConfigI, _ := meta.GetRawConfigForKey(d, "private_network.#."+ipFieldName, cty.String); staticConfigI != nil {
		staticConfig = types.ExpandStringPtr(staticConfigI)
	}

	return ipamConfig, staticConfig
}

// LoadBalancerDiffSuppressFunc suppresses diff when load_balancer is not set
func LoadBalancerDiffSuppressFunc(k, oldValue, newValue string, d *schema.ResourceData) bool {
	if !strings.HasPrefix(k, "load_balancer") {
		return false
	}

	if _, exists := d.GetOk("load_balancer"); !exists {
		return true
	}

	return false
}

// retryRDBReadOnTransient wraps a read action to handle transient_state (HTTP 409) by
// waiting for the instance to be ready, then retrying the action.
func retryRDBReadOnTransient[T any](ctx context.Context, api *rdb.API, region scw.Region, instanceID string, action func() (T, error)) (T, error) {
	return transport.RetryOnTransientStateError(
		action,
		func() (*rdb.Instance, error) {
			return waitForRDBInstance(ctx, api, region, instanceID, defaultInstanceTimeout)
		},
	)
}

// applyInstanceSettings merges user settings onto the instance defaults then calls SetInstanceSettings.
// Unlike a bare Set of the config map, this preserves engine defaults not listed in Terraform.
func applyInstanceSettings(ctx context.Context, api *rdb.API, region scw.Region, instanceID string, timeout time.Duration, oldManaged, newManaged map[string]string) error {
	res, err := waitForRDBInstance(ctx, api, region, instanceID, timeout)
	if err != nil {
		return err
	}

	current, ok := flattenInstanceSettings(res.Settings).(map[string]string)
	if !ok {
		return errors.New("unexpected type for instance settings")
	}

	// Legacy state mirrored all API defaults as "settings"; treat that as unmanaged.
	if maps.Equal(oldManaged, current) {
		oldManaged = nil
	}

	merged := MergeInstanceSettings(current, oldManaged, newManaged)

	_, err = api.SetInstanceSettings(&rdb.SetInstanceSettingsRequest{
		InstanceID: instanceID,
		Region:     region,
		Settings:   expandInstanceSettingsFromMap(merged),
	}, scw.WithContext(ctx))

	return err
}

func rawConfigSettingsKeys(d *schema.ResourceData) (map[string]bool, bool) {
	raw := d.GetRawConfig()
	if raw.IsNull() || !raw.IsKnown() {
		return nil, false
	}

	attr := raw.GetAttr("settings")
	if attr.IsNull() {
		return nil, false
	}

	valueMap := attr.AsValueMap()
	if len(valueMap) == 0 {
		return nil, false
	}

	keys := make(map[string]bool, len(valueMap))
	for key := range valueMap {
		keys[key] = true
	}

	return keys, true
}

// setInstanceSettingsState persists only HCL-managed settings keys into state.
func setInstanceSettingsState(d *schema.ResourceData, settings []*rdb.InstanceSetting) {
	allSettings, ok := flattenInstanceSettings(settings).(map[string]string)
	if !ok {
		allSettings = map[string]string{}
	}

	configKeys, managed := rawConfigSettingsKeys(d)
	raw := d.GetRawConfig()
	rawKnown := !raw.IsNull() && raw.IsKnown()

	switch {
	case managed:
		_ = d.Set("settings", FilterInstanceSettings(allSettings, configKeys))
	case rawKnown:
		// HCL is known and has no settings block — clear legacy computed defaults from state.
		_ = d.Set("settings", nil)
	default:
		// Raw config unavailable (some create Read paths): fall back to GetOk keys only.
		if v, ok := d.GetOk("settings"); ok {
			fallbackKeys := make(map[string]bool)
			for key := range v.(map[string]any) {
				fallbackKeys[key] = true
			}

			_ = d.Set("settings", FilterInstanceSettings(allSettings, fallbackKeys))
		} else {
			_ = d.Set("settings", nil)
		}
	}
}

func isTimeoutErr(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	msg := strings.ToLower(err.Error())

	return strings.Contains(msg, "timeout") || strings.Contains(msg, "deadline exceeded")
}

// majorUpgradeTimeoutOrErr returns a clear diagnostic when a blue/green engine upgrade
// wait times out. The Scaleway workflow may still continue after Terraform gives up.
func majorUpgradeTimeoutOrErr(err error, region scw.Region, newInstanceID, oldInstanceID string) diag.Diagnostics {
	if !isTimeoutErr(err) {
		return diag.FromErr(err)
	}

	newID := regional.NewIDString(region, newInstanceID)
	oldID := regional.NewIDString(region, oldInstanceID)

	return diag.Diagnostics{{
		Severity: diag.Error,
		Summary:  "RDB engine upgrade timed out",
		Detail: fmt.Sprintf(
			"Terraform timed out while waiting for the blue/green engine upgrade to finish. "+
				"The Scaleway upgrade may still be running in the background (snapshot/restore/endpoint migration).\n\n"+
				"New instance: %s\nOld instance: %s\n\n"+
				"Check both instances (status, engine, endpoints), e.g. `scw rdb instance get %s region=%s`. "+
				"If the new instance is ready with the target engine and endpoints migrated, ensure Terraform state points to the new ID and delete the old instance manually if it remains. "+
				"Do not re-apply an engine change while an upgrade is still in progress: "+
				"if state still points to the old instance, Terraform may start another blue/green upgrade and create more orphaned instances. "+
				"For large or HA upgrades, increase timeouts.update (default is 60m).\n\nUnderlying error: %v",
			newID, oldID, newInstanceID, region, err,
		),
	}}
}
