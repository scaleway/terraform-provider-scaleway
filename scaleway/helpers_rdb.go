package scaleway

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/ipam/v1"
	"github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

const (
	defaultRdbInstanceTimeout   = 15 * time.Minute
	defaultWaitRDBRetryInterval = 30 * time.Second
)

// newRdbAPI returns a new RDB API
func newRdbAPI(m interface{}) *rdb.API {
	return rdb.NewAPI(meta.ExtractScwClient(m))
}

// rdbAPIWithRegion returns a new lb API and the region for a Create request
func rdbAPIWithRegion(d *schema.ResourceData, m interface{}) (*rdb.API, scw.Region, error) {
	region, err := meta.ExtractRegion(d, m)
	if err != nil {
		return nil, "", err
	}
	return newRdbAPI(m), region, nil
}

// rdbAPIWithRegionAndID returns an lb API with region and ID extracted from the state
func rdbAPIWithRegionAndID(m interface{}, id string) (*rdb.API, scw.Region, string, error) {
	region, ID, err := regional.ParseID(id)
	if err != nil {
		return nil, "", "", err
	}
	return newRdbAPI(m), region, ID, nil
}

func flattenInstanceSettings(settings []*rdb.InstanceSetting) interface{} {
	res := make(map[string]string)
	for _, value := range settings {
		res[value.Name] = value.Value
	}

	return res
}

func expandInstanceSettings(i interface{}) []*rdb.InstanceSetting {
	rawRule := i.(map[string]interface{})
	res := make([]*rdb.InstanceSetting, 0, len(rawRule))
	for key, value := range rawRule {
		res = append(res, &rdb.InstanceSetting{
			Name:  key,
			Value: value.(string),
		})
	}

	return res
}

func waitForRDBInstance(ctx context.Context, api *rdb.API, region scw.Region, id string, timeout time.Duration) (*rdb.Instance, error) {
	retryInterval := defaultWaitRDBRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	return api.WaitForInstance(&rdb.WaitForInstanceRequest{
		Region:        region,
		Timeout:       scw.TimeDurationPtr(timeout),
		InstanceID:    id,
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))
}

func waitForRDBDatabaseBackup(ctx context.Context, api *rdb.API, region scw.Region, id string, timeout time.Duration) (*rdb.DatabaseBackup, error) {
	retryInterval := defaultWaitRDBRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	return api.WaitForDatabaseBackup(&rdb.WaitForDatabaseBackupRequest{
		Region:           region,
		Timeout:          scw.TimeDurationPtr(timeout),
		DatabaseBackupID: id,
		RetryInterval:    &retryInterval,
	}, scw.WithContext(ctx))
}

func waitForRDBReadReplica(ctx context.Context, api *rdb.API, region scw.Region, id string, timeout time.Duration) (*rdb.ReadReplica, error) {
	retryInterval := defaultWaitRDBRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	return api.WaitForReadReplica(&rdb.WaitForReadReplicaRequest{
		Region:        region,
		Timeout:       scw.TimeDurationPtr(timeout),
		ReadReplicaID: id,
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))
}

func expandPrivateNetwork(data interface{}, exist bool, ipamConfig *bool, staticConfig *string) ([]*rdb.EndpointSpec, diag.Diagnostics) {
	if data == nil || !exist {
		return nil, nil
	}
	var diags diag.Diagnostics

	res := make([]*rdb.EndpointSpec, 0, len(data.([]interface{})))
	for _, pn := range data.([]interface{}) {
		r := pn.(map[string]interface{})
		spec := &rdb.EndpointSpec{
			PrivateNetwork: &rdb.EndpointSpecPrivateNetwork{
				PrivateNetworkID: locality.ExpandID(r["pn_id"].(string)),
				IpamConfig:       &rdb.EndpointSpecPrivateNetworkIpamConfig{},
			},
		}

		if staticConfig != nil {
			ip, err := expandIPNet(*staticConfig)
			if err != nil {
				return nil, append(diags, diag.FromErr(fmt.Errorf("failed to parse private_network ip_net (%s): %s", r["ip_net"], err))...)
			}
			spec.PrivateNetwork.ServiceIP = &ip
			spec.PrivateNetwork.IpamConfig = nil
			if ipamConfig != nil && *ipamConfig {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Warning,
					Detail:   "`ip_net` field is set so `enable_ipam` field will be ignored",
				})
			}
		} else if ipamConfig == nil || !*ipamConfig {
			return nil, diag.FromErr(errors.New("at least one of `ip_net` or `enable_ipam` (set to true) must be set"))
		}
		res = append(res, spec)
	}

	return res, diags
}

func expandLoadBalancer() *rdb.EndpointSpec {
	return &rdb.EndpointSpec{
		LoadBalancer: &rdb.EndpointSpecLoadBalancer{},
	}
}

func flattenPrivateNetwork(endpoints []*rdb.Endpoint, enableIpam bool) (interface{}, bool) {
	pnI := []map[string]interface{}(nil)
	for _, endpoint := range endpoints {
		if endpoint.PrivateNetwork != nil {
			pn := endpoint.PrivateNetwork
			fetchRegion, err := pn.Zone.Region()
			if err != nil {
				return diag.FromErr(err), false
			}
			pnRegionalID := regional.NewIDString(fetchRegion, pn.PrivateNetworkID)
			serviceIP, err := flattenIPNet(pn.ServiceIP)
			if err != nil {
				return pnI, false
			}
			pnI = append(pnI, map[string]interface{}{
				"endpoint_id": endpoint.ID,
				"ip":          flattenIPPtr(endpoint.IP),
				"port":        int(endpoint.Port),
				"name":        endpoint.Name,
				"ip_net":      serviceIP,
				"pn_id":       pnRegionalID,
				"hostname":    flattenStringPtr(endpoint.Hostname),
				"enable_ipam": enableIpam,
			})
			return pnI, true
		}
	}

	return pnI, false
}

func flattenLoadBalancer(endpoints []*rdb.Endpoint) (interface{}, bool) {
	flat := []map[string]interface{}(nil)
	for _, endpoint := range endpoints {
		if endpoint.LoadBalancer != nil {
			flat = append(flat, map[string]interface{}{
				"endpoint_id": endpoint.ID,
				"ip":          flattenIPPtr(endpoint.IP),
				"port":        int(endpoint.Port),
				"name":        endpoint.Name,
				"hostname":    flattenStringPtr(endpoint.Hostname),
			})
			return flat, true
		}
	}

	return flat, false
}

// expandTimePtr returns a time pointer for an RFC3339 time.
// It returns nil if time is not valid, you should use validateDate to validate field.
func expandTimePtr(i interface{}) *time.Time {
	rawTime := expandStringPtr(i)
	if rawTime == nil {
		return nil
	}
	parsedTime, err := time.Parse(time.RFC3339, *rawTime)
	if err != nil {
		return nil
	}
	return &parsedTime
}

func expandReadReplicaEndpointsSpecDirectAccess(data interface{}) *rdb.ReadReplicaEndpointSpec {
	if data == nil || len(data.([]interface{})) == 0 {
		return nil
	}

	return &rdb.ReadReplicaEndpointSpec{
		DirectAccess: new(rdb.ReadReplicaEndpointSpecDirectAccess),
	}
}

// expandReadReplicaEndpointsSpecPrivateNetwork expand read-replica private network endpoints from schema to specs
func expandReadReplicaEndpointsSpecPrivateNetwork(data interface{}, ipamConfig *bool, staticConfig *string) (*rdb.ReadReplicaEndpointSpec, diag.Diagnostics) {
	if data == nil || len(data.([]interface{})) == 0 {
		return nil, nil
	}
	// private_network is a list of size 1
	data = data.([]interface{})[0]

	rawEndpoint := data.(map[string]interface{})
	var diags diag.Diagnostics

	endpoint := &rdb.ReadReplicaEndpointSpec{
		PrivateNetwork: &rdb.ReadReplicaEndpointSpecPrivateNetwork{
			PrivateNetworkID: locality.ExpandID(rawEndpoint["private_network_id"]),
			IpamConfig:       &rdb.ReadReplicaEndpointSpecPrivateNetworkIpamConfig{},
		},
	}

	if staticConfig != nil {
		ipNet, err := expandIPNet(*staticConfig)
		if err != nil {
			return nil, append(diags, diag.FromErr(fmt.Errorf("failed to parse private_network service_ip (%s): %s", rawEndpoint["service_ip"], err))...)
		}
		endpoint.PrivateNetwork.ServiceIP = &ipNet
		endpoint.PrivateNetwork.IpamConfig = nil
		if ipamConfig != nil && !*ipamConfig {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Detail:   "`service_ip` field is set so `enable_ipam` field will be ignored",
			})
		}
	} else if ipamConfig == nil || !*ipamConfig {
		return nil, diag.FromErr(errors.New("at least one of `service_ip` or `enable_ipam` (set to true) must be set"))
	}

	return endpoint, diags
}

// flattenReadReplicaEndpoints flatten read-replica endpoints to directAccess and privateNetwork
func flattenReadReplicaEndpoints(endpoints []*rdb.Endpoint, enableIpam bool) (directAccess, privateNetwork interface{}) {
	for _, endpoint := range endpoints {
		rawEndpoint := map[string]interface{}{
			"endpoint_id": endpoint.ID,
			"ip":          flattenIPPtr(endpoint.IP),
			"port":        int(endpoint.Port),
			"name":        endpoint.Name,
			"hostname":    flattenStringPtr(endpoint.Hostname),
		}
		if endpoint.DirectAccess != nil {
			directAccess = rawEndpoint
		}
		if endpoint.PrivateNetwork != nil {
			fetchRegion, err := endpoint.PrivateNetwork.Zone.Region()
			if err != nil {
				return diag.FromErr(err), false
			}
			pnRegionalID := regional.NewIDString(fetchRegion, endpoint.PrivateNetwork.PrivateNetworkID)
			rawEndpoint["private_network_id"] = pnRegionalID
			rawEndpoint["service_ip"] = endpoint.PrivateNetwork.ServiceIP.String()
			rawEndpoint["zone"] = endpoint.PrivateNetwork.Zone
			rawEndpoint["enable_ipam"] = enableIpam
			privateNetwork = rawEndpoint
		}
	}

	// direct_access and private_network are lists

	if directAccess != nil {
		directAccess = []interface{}{directAccess}
	}
	if privateNetwork != nil {
		privateNetwork = []interface{}{privateNetwork}
	}

	return directAccess, privateNetwork
}

// rdbPrivilegeV1SchemaUpgradeFunc allow upgrade the privilege ID on schema V1
func rdbPrivilegeV1SchemaUpgradeFunc(_ context.Context, rawState map[string]interface{}, m interface{}) (map[string]interface{}, error) {
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
	rawState["id"] = resourceScalewayRdbUserPrivilegeID(region, idStr, databaseName, userName)
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
		ipamConfig = expandBoolPtr(enableIpam)
	}
	customIP, customIPSet := d.GetOk("private_network.0." + ipFieldName)
	if customIPSet {
		staticConfig = expandStringPtr(customIP)
	}
	return ipamConfig, staticConfig
}

// getIPConfigUpdate forces the provider to read the user's config instead of checking the state, because "enable_ipam" is not readable from the API
func getIPConfigUpdate(d *schema.ResourceData, ipFieldName string) (ipamConfig *bool, staticConfig *string) {
	if rawConfig := d.GetRawConfig(); !rawConfig.IsNull() {
		pnRawConfig := rawConfig.AsValueMap()["private_network"].AsValueSlice()[0].AsValueMap()
		if !pnRawConfig["enable_ipam"].IsNull() {
			if pnRawConfig["enable_ipam"].False() {
				ipamConfig = scw.BoolPtr(false)
			} else {
				ipamConfig = scw.BoolPtr(true)
			}
		}
		if !pnRawConfig[ipFieldName].IsNull() {
			value := pnRawConfig[ipFieldName].AsString()
			staticConfig = &value
		}
	}
	return ipamConfig, staticConfig
}

func getIPAMConfigRead(resource interface{}, m interface{}) (bool, error) {
	ipamAPI := ipam.NewAPI(meta.ExtractScwClient(m))
	request := &ipam.ListIPsRequest{
		ResourceType: "rdb_instance",
		IsIPv6:       scw.BoolPtr(false),
	}
	var privateEndpoint *rdb.EndpointPrivateNetworkDetails

	switch res := resource.(type) {
	case *rdb.Instance:
		request.Region = res.Region
		request.ResourceID = &res.ID
		for _, e := range res.Endpoints {
			if e.PrivateNetwork != nil {
				privateEndpoint = e.PrivateNetwork
			}
		}
	case *rdb.ReadReplica:
		request.Region = res.Region
		request.ResourceID = &res.InstanceID
		for _, e := range res.Endpoints {
			if e.PrivateNetwork != nil {
				privateEndpoint = e.PrivateNetwork
			}
		}
	}
	if privateEndpoint == nil {
		return false, nil
	}

	ips, err := ipamAPI.ListIPs(request, scw.WithAllPages())
	if err != nil {
		return false, fmt.Errorf("could not list IPs: %w", err)
	}

	for _, ip := range ips.IPs {
		if ip.Address.String() == privateEndpoint.ServiceIP.String() {
			return true, nil
		}
	}
	return false, nil
}
