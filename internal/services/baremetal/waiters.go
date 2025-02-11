package baremetal

import (
	"context"
	"time"

	"github.com/scaleway/scaleway-sdk-go/api/baremetal/v1"
	baremetalV3 "github.com/scaleway/scaleway-sdk-go/api/baremetal/v3"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

func waitForServer(ctx context.Context, api *baremetal.API, zone scw.Zone, serverID string, timeout time.Duration) (*baremetal.Server, error) {
	retryInterval := retryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	server, err := api.WaitForServer(&baremetal.WaitForServerRequest{
		Zone:          zone,
		ServerID:      serverID,
		Timeout:       scw.TimeDurationPtr(timeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))

	return server, err
}

func waitForServerInstall(ctx context.Context, api *baremetal.API, zone scw.Zone, serverID string, timeout time.Duration) (*baremetal.Server, error) {
	retryInterval := retryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	server, err := api.WaitForServerInstall(&baremetal.WaitForServerInstallRequest{
		Zone:          zone,
		ServerID:      serverID,
		Timeout:       scw.TimeDurationPtr(timeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))

	return server, err
}

func waitForServerOptions(ctx context.Context, api *baremetal.API, zone scw.Zone, serverID string, timeout time.Duration) (*baremetal.Server, error) {
	retryInterval := retryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	server, err := api.WaitForServerOptions(&baremetal.WaitForServerOptionsRequest{
		Zone:          zone,
		ServerID:      serverID,
		Timeout:       scw.TimeDurationPtr(timeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))

	return server, err
}

func waitForServerPrivateNetwork(ctx context.Context, api *baremetalV3.PrivateNetworkAPI, zone scw.Zone, serverID string, timeout time.Duration) ([]*baremetalV3.ServerPrivateNetwork, error) {
	retryInterval := retryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	serverPrivateNetwork, err := api.WaitForServerPrivateNetworks(&baremetalV3.WaitForServerPrivateNetworksRequest{
		Zone:          zone,
		ServerID:      serverID,
		Timeout:       scw.TimeDurationPtr(timeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))

	return serverPrivateNetwork, err
}
