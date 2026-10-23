package vpcgw

import (
	"context"
	"fmt"
	"time"

	v2 "github.com/scaleway/scaleway-sdk-go/api/vpcgw/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

func waitForVPCPublicGateway(ctx context.Context, api *v2.API, zone scw.Zone, id string, timeout time.Duration) (*v2.Gateway, error) {
	retryInterval := defaultRetry
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	gateway, err := api.WaitForGateway(&v2.WaitForGatewayRequest{
		Timeout:       new(timeout),
		GatewayID:     id,
		RetryInterval: &retryInterval,
		Zone:          zone,
	}, scw.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("waiting for vpcgw gateway to be ready: %w", err)
	}

	return gateway, nil
}

func waitForVPCGatewayNetwork(ctx context.Context, api *v2.API, zone scw.Zone, id string, timeout time.Duration) (*v2.GatewayNetwork, error) {
	retryIntervalGWNetwork := defaultRetry
	if transport.DefaultWaitRetryInterval != nil {
		retryIntervalGWNetwork = *transport.DefaultWaitRetryInterval
	}

	gatewayNetwork, err := api.WaitForGatewayNetwork(&v2.WaitForGatewayNetworkRequest{
		GatewayNetworkID: id,
		Timeout:          new(timeout),
		RetryInterval:    &retryIntervalGWNetwork,
		Zone:             zone,
	}, scw.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("waiting for vpcgw gateway network to be ready: %w", err)
	}

	return gatewayNetwork, nil
}
