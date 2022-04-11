package scaleway

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	vpcgw "github.com/scaleway/scaleway-sdk-go/api/vpcgw/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	defaultVPCGatewayTimeout = 10 * time.Minute
	defaultVPCGatewayRetry   = 5 * time.Second
)

// vpcgwAPIWithZone returns a new VPC API and the zone for a Create request
func vpcgwAPIWithZone(d *schema.ResourceData, m interface{}) (*vpcgw.API, scw.Zone, error) {
	meta := m.(*Meta)
	vpcgwAPI := vpcgw.NewAPI(meta.scwClient)

	zone, err := extractZone(d, meta)
	if err != nil {
		return nil, "", err
	}
	return vpcgwAPI, zone, nil
}

// vpcgwAPIWithZoneAndID
func vpcgwAPIWithZoneAndID(m interface{}, id string) (*vpcgw.API, scw.Zone, string, error) {
	meta := m.(*Meta)
	vpcgwAPI := vpcgw.NewAPI(meta.scwClient)

	zone, ID, err := parseZonedID(id)
	if err != nil {
		return nil, "", "", err
	}
	return vpcgwAPI, zone, ID, nil
}

func waitForVPCPublicGateway(ctx context.Context, api *vpcgw.API, zone scw.Zone, id string, timeout time.Duration) (*vpcgw.Gateway, error) {
	retryInterval := defaultVPCGatewayRetry
	if DefaultWaitRetryInterval != nil {
		retryInterval = *DefaultWaitRetryInterval
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
	if DefaultWaitRetryInterval != nil {
		retryIntervalGWNetwork = *DefaultWaitRetryInterval
	}

	gwNetwork, err := api.WaitForGatewayNetwork(&vpcgw.WaitForGatewayNetworkRequest{
		GatewayNetworkID: id,
		Timeout:          scw.TimeDurationPtr(timeout),
		RetryInterval:    &retryIntervalGWNetwork,
		Zone:             zone,
	}, scw.WithContext(ctx))

	return gwNetwork, err
}
