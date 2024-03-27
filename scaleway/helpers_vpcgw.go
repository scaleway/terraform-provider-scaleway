package scaleway

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	vpcgw "github.com/scaleway/scaleway-sdk-go/api/vpcgw/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

const (
	defaultVPCGatewayTimeout                   = 10 * time.Minute
	defaultVPCGatewayRetry                     = 5 * time.Second
	defaultVPCPublicGatewayIPReverseDNSTimeout = 10 * time.Minute
)

// vpcgwAPIWithZone returns a new VPC API and the zone for a Create request
func vpcgwAPIWithZone(d *schema.ResourceData, m interface{}) (*vpcgw.API, scw.Zone, error) {
	vpcgwAPI := vpcgw.NewAPI(meta.ExtractScwClient(m))

	zone, err := meta.ExtractZone(d, m)
	if err != nil {
		return nil, "", err
	}
	return vpcgwAPI, zone, nil
}

// VpcgwAPIWithZoneAndID
func VpcgwAPIWithZoneAndID(m interface{}, id string) (*vpcgw.API, scw.Zone, string, error) {
	vpcgwAPI := vpcgw.NewAPI(meta.ExtractScwClient(m))

	zone, ID, err := zonal.ParseID(id)
	if err != nil {
		return nil, "", "", err
	}
	return vpcgwAPI, zone, ID, nil
}

func waitForVPCPublicGateway(ctx context.Context, api *vpcgw.API, zone scw.Zone, id string, timeout time.Duration) (*vpcgw.Gateway, error) {
	retryInterval := defaultVPCGatewayRetry
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	gateway, err := api.WaitForGateway(&vpcgw.WaitForGatewayRequest{
		Timeout:       scw.TimeDurationPtr(timeout),
		GatewayID:     id,
		RetryInterval: &retryInterval,
		Zone:          zone,
	}, scw.WithContext(ctx))

	return gateway, err
}

func waitForVPCGatewayNetwork(ctx context.Context, api *vpcgw.API, zone scw.Zone, id string, timeout time.Duration) (*vpcgw.GatewayNetwork, error) {
	retryIntervalGWNetwork := defaultVPCGatewayRetry
	if transport.DefaultWaitRetryInterval != nil {
		retryIntervalGWNetwork = *transport.DefaultWaitRetryInterval
	}

	gatewayNetwork, err := api.WaitForGatewayNetwork(&vpcgw.WaitForGatewayNetworkRequest{
		GatewayNetworkID: id,
		Timeout:          scw.TimeDurationPtr(timeout),
		RetryInterval:    &retryIntervalGWNetwork,
		Zone:             zone,
	}, scw.WithContext(ctx))

	return gatewayNetwork, err
}

func waitForDHCPEntries(ctx context.Context, api *vpcgw.API, zone scw.Zone, gatewayID string, macAddress string, timeout time.Duration) (*vpcgw.ListDHCPEntriesResponse, error) {
	retryIntervalDHCPEntries := defaultVPCGatewayRetry
	if transport.DefaultWaitRetryInterval != nil {
		retryIntervalDHCPEntries = *transport.DefaultWaitRetryInterval
	}

	req := &vpcgw.WaitForDHCPEntriesRequest{
		MacAddress:    macAddress,
		Zone:          zone,
		Timeout:       scw.TimeDurationPtr(timeout),
		RetryInterval: &retryIntervalDHCPEntries,
	}

	if gatewayID != "" {
		req.GatewayNetworkID = &gatewayID
	}

	dhcpEntries, err := api.WaitForDHCPEntries(req, scw.WithContext(ctx))
	return dhcpEntries, err
}

func retryUpdateGatewayReverseDNS(ctx context.Context, api *vpcgw.API, req *vpcgw.UpdateIPRequest, timeout time.Duration) error {
	timeoutChannel := time.After(timeout)

	for {
		select {
		case <-time.After(defaultVPCGatewayRetry):
			_, err := api.UpdateIP(req, scw.WithContext(ctx))
			if err != nil && isIPReverseDNSResolveError(err) {
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
