package vpcgw

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/vpcgw/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/instance"
)

const (
	defaultTimeout             = 10 * time.Minute
	defaultRetry               = 5 * time.Second
	defaultIPReverseDNSTimeout = 10 * time.Minute
)

// newAPIWithZone returns a new VPC API and the zone for a Create request
func newAPIWithZone(d *schema.ResourceData, m interface{}) (*vpcgw.API, scw.Zone, error) {
	api := vpcgw.NewAPI(meta.ExtractScwClient(m))

	zone, err := meta.ExtractZone(d, m)
	if err != nil {
		return nil, "", err
	}

	return api, zone, nil
}

func NewAPIWithZoneAndID(m interface{}, id string) (*vpcgw.API, scw.Zone, string, error) {
	api := vpcgw.NewAPI(meta.ExtractScwClient(m))

	zone, ID, err := zonal.ParseID(id)
	if err != nil {
		return nil, "", "", err
	}

	return api, zone, ID, nil
}

func retryUpdateGatewayReverseDNS(ctx context.Context, api *vpcgw.API, req *vpcgw.UpdateIPRequest, timeout time.Duration) error {
	timeoutChannel := time.After(timeout)

	for {
		select {
		case <-time.After(defaultRetry):
			_, err := api.UpdateIP(req, scw.WithContext(ctx))
			if err != nil && instance.IsIPReverseDNSResolveError(err) {
				continue
			}

			return err
		case <-timeoutChannel:
			_, err := api.UpdateIP(req, scw.WithContext(ctx))

			return err
		}
	}
}

func expandIpamConfig(raw interface{}) *vpcgw.CreateGatewayNetworkRequestIpamConfig {
	if raw == nil || len(raw.([]interface{})) != 1 {
		return nil
	}

	rawMap := raw.([]interface{})[0].(map[string]interface{})

	ipamConfig := &vpcgw.CreateGatewayNetworkRequestIpamConfig{
		PushDefaultRoute: rawMap["push_default_route"].(bool),
	}

	if ipamIPID, ok := rawMap["ipam_ip_id"].(string); ok && ipamIPID != "" {
		ipamConfig.IpamIPID = scw.StringPtr(regional.ExpandID(ipamIPID).ID)
	}

	return ipamConfig
}

func expandUpdateIpamConfig(raw interface{}) *vpcgw.UpdateGatewayNetworkRequestIpamConfig {
	if raw == nil || len(raw.([]interface{})) != 1 {
		return nil
	}

	rawMap := raw.([]interface{})[0].(map[string]interface{})

	updateIpamConfig := &vpcgw.UpdateGatewayNetworkRequestIpamConfig{
		PushDefaultRoute: scw.BoolPtr(rawMap["push_default_route"].(bool)),
	}

	if ipamIPID, ok := rawMap["ipam_ip_id"].(string); ok && ipamIPID != "" {
		updateIpamConfig.IpamIPID = scw.StringPtr(regional.ExpandID(ipamIPID).ID)
	}

	return updateIpamConfig
}

func flattenIpamConfig(config *vpcgw.IpamConfig) interface{} {
	if config == nil {
		return nil
	}

	return []map[string]interface{}{
		{
			"push_default_route": config.PushDefaultRoute,
			"ipam_ip_id":         config.IpamIPID,
		},
	}
}
