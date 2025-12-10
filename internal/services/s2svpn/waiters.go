package s2svpn

import (
	"context"
	"time"

	s2s_vpn "github.com/scaleway/scaleway-sdk-go/api/s2s_vpn/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

const (
	defaulVPNGatewayTimeout        = 5 * time.Minute
	defaultVPNGatewayRetryInterval = 5 * time.Second
)

func waitForVPNGateway(ctx context.Context, api *s2s_vpn.API, region scw.Region, vpngwID string, timeout time.Duration) (*s2s_vpn.VpnGateway, error) {
	retryInterval := defaultVPNGatewayRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	server, err := api.WaitForVpnGateway(&s2s_vpn.WaitForVpnGatewayRequest{
		GatewayID:     vpngwID,
		Region:        region,
		Timeout:       scw.TimeDurationPtr(timeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))

	return server, err
}
