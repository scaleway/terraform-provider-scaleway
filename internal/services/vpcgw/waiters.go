package vpcgw

import (
	"context"
	"time"

	"github.com/scaleway/scaleway-sdk-go/api/vpcgw/v1"
	v2 "github.com/scaleway/scaleway-sdk-go/api/vpcgw/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

func waitForVPCPublicGateway(ctx context.Context, api *vpcgw.API, zone scw.Zone, id string, timeout time.Duration) (*vpcgw.Gateway, error) {
	retryInterval := defaultRetry
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

func waitForVPCPublicGatewayV2(ctx context.Context, api *v2.API, zone scw.Zone, id string, timeout time.Duration) (*v2.Gateway, error) {
	retryInterval := defaultRetry
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	gateway, err := api.WaitForGateway(&v2.WaitForGatewayRequest{
		Timeout:       scw.TimeDurationPtr(timeout),
		GatewayID:     id,
		RetryInterval: &retryInterval,
		Zone:          zone,
	}, scw.WithContext(ctx))

	return gateway, err
}

func waitForVPCGatewayNetwork(ctx context.Context, api *vpcgw.API, zone scw.Zone, id string, timeout time.Duration) (*vpcgw.GatewayNetwork, error) {
	retryIntervalGWNetwork := defaultRetry
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

func waitForVPCGatewayNetworkV2(ctx context.Context, api *v2.API, zone scw.Zone, id string, timeout time.Duration) (*v2.GatewayNetwork, error) {
	retryIntervalGWNetwork := defaultRetry
	if transport.DefaultWaitRetryInterval != nil {
		retryIntervalGWNetwork = *transport.DefaultWaitRetryInterval
	}

	gatewayNetwork, err := api.WaitForGatewayNetwork(&v2.WaitForGatewayNetworkRequest{
		GatewayNetworkID: id,
		Timeout:          scw.TimeDurationPtr(timeout),
		RetryInterval:    &retryIntervalGWNetwork,
		Zone:             zone,
	}, scw.WithContext(ctx))

	return gatewayNetwork, err
}

func waitForDHCPEntries(ctx context.Context, api *vpcgw.API, zone scw.Zone, gatewayID string, macAddress string, timeout time.Duration) (*vpcgw.ListDHCPEntriesResponse, error) {
	retryIntervalDHCPEntries := defaultRetry
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
