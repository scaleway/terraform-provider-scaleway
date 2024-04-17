package rdb

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/ipam/v1"
	"github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

const (
	defaultInstanceTimeout           = 15 * time.Minute
	defaultWaitRetryInterval         = 30 * time.Second
	defaultMaxAgeRetention    uint32 = 30
	defaultTotalDiskRetention        = scw.Size(0)
)

// newAPI returns a new RDB API
func newAPI(m interface{}) *rdb.API {
	return rdb.NewAPI(meta.ExtractScwClient(m))
}

// newAPIWithRegion returns a new lb API and the region for a Create request
func newAPIWithRegion(d *schema.ResourceData, m interface{}) (*rdb.API, scw.Region, error) {
	region, err := meta.ExtractRegion(d, m)
	if err != nil {
		return nil, "", err
	}
	return newAPI(m), region, nil
}

// NewAPIWithRegionAndID returns an lb API with region and ID extracted from the state
func NewAPIWithRegionAndID(m interface{}, id string) (*rdb.API, scw.Region, string, error) {
	region, ID, err := regional.ParseID(id)
	if err != nil {
		return nil, "", "", err
	}
	return newAPI(m), region, ID, nil
}

// PrivilegeV1SchemaUpgradeFunc allow upgrade the privilege ID on schema V1
func PrivilegeV1SchemaUpgradeFunc(_ context.Context, rawState map[string]interface{}, m interface{}) (map[string]interface{}, error) {
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

func flattenSizePtr(i *scw.Size) interface{} {
	if i == nil {
		return 0
	}
	return *i
}

func expandSizePtr(data interface{}) *scw.Size {
	if data == nil || data == "" {
		return nil
	}

	size := scw.Size(data.(int))
	return &size
}

func expandInstanceLogsPolicy(i interface{}) *rdb.LogsPolicy {
	policyConfigRaw := i.([]interface{})
	for _, policyRaw := range policyConfigRaw {
		policy := policyRaw.(map[string]interface{})
		return &rdb.LogsPolicy{
			MaxAgeRetention:    types.ExpandUint32Ptr(policy["max_age_retention"]),
			TotalDiskRetention: expandSizePtr(policy["total_disk_retention"]),
		}
	}
	return nil
}

func flattenInstanceLogsPolicy(policy *rdb.LogsPolicy) interface{} {
	p := []map[string]interface{}{}
	if policy != nil {
		p = append(p, map[string]interface{}{
			"max_age_retention":    types.FlattenUint32Ptr(policy.MaxAgeRetention),
			"total_disk_retention": flattenSizePtr(policy.TotalDiskRetention),
		})
	} else {
		return nil
	}
	return p
}
