package scaleway

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	vpcgw "github.com/scaleway/scaleway-sdk-go/api/vpcgw/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	gatewayWaitForTimeout    = 10 * time.Minute
	defaultVPCGatewayTimeout = 10 * time.Minute
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

type vpcHandler struct {
	ctx  context.Context
	api  *vpcgw.API
	zone scw.Zone
}

func newVPCHandler(ctx context.Context, api *vpcgw.API, zone scw.Zone) *vpcHandler {
	return &vpcHandler{ctx: ctx, api: api, zone: zone}
}

func (vpcH *vpcHandler) waitGNetwork(gwNetworkID string) (*vpcgw.GatewayNetwork, error) {
	retryInterval := retryIntervalVPCGatewayNetwork
	return vpcH.api.WaitForGatewayNetwork(&vpcgw.WaitForGatewayNetworkRequest{
		GatewayNetworkID: gwNetworkID,
		Timeout:          scw.TimeDurationPtr(defaultVPCGatewayTimeout),
		RetryInterval:    &retryInterval,
		Zone:             vpcH.zone,
	}, scw.WithContext(vpcH.ctx))
}

func (vpcH *vpcHandler) waitGateway(gwID string) (*vpcgw.Gateway, error) {
	retryInterval := retryGWTimeout
	return vpcH.api.WaitForGateway(&vpcgw.WaitForGatewayRequest{
		GatewayID:     gwID,
		Timeout:       scw.TimeDurationPtr(defaultVPCGatewayTimeout),
		RetryInterval: &retryInterval,
		Zone:          vpcH.zone,
	}, scw.WithContext(vpcH.ctx))
}
