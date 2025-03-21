package instance

import (
	"context"
	"time"

	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/instance/instancehelpers"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

func waitForSnapshot(ctx context.Context, api *instance.API, zone scw.Zone, id string, timeout time.Duration) (*instance.Snapshot, error) {
	retryInterval := instancehelpers.DefaultInstanceRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	snapshot, err := api.WaitForSnapshot(&instance.WaitForSnapshotRequest{
		SnapshotID:    id,
		Zone:          zone,
		Timeout:       scw.TimeDurationPtr(timeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))

	return snapshot, err
}

func waitForServer(ctx context.Context, api *instance.API, zone scw.Zone, id string, timeout time.Duration) (*instance.Server, error) {
	retryInterval := instancehelpers.DefaultInstanceRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	server, err := api.WaitForServer(&instance.WaitForServerRequest{
		Zone:          zone,
		ServerID:      id,
		Timeout:       scw.TimeDurationPtr(timeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))

	return server, err
}

func waitForPrivateNIC(ctx context.Context, instanceAPI *instance.API, zone scw.Zone, serverID string, privateNICID string, timeout time.Duration) (*instance.PrivateNIC, error) {
	retryInterval := instancehelpers.DefaultInstanceRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	nic, err := instanceAPI.WaitForPrivateNIC(&instance.WaitForPrivateNICRequest{
		ServerID:      serverID,
		PrivateNicID:  privateNICID,
		Zone:          zone,
		Timeout:       scw.TimeDurationPtr(timeout),
		RetryInterval: scw.TimeDurationPtr(retryInterval),
	}, scw.WithContext(ctx))

	return nic, err
}

func waitForMACAddress(ctx context.Context, instanceAPI *instance.API, zone scw.Zone, serverID string, privateNICID string, timeout time.Duration) (*instance.PrivateNIC, error) {
	retryInterval := instancehelpers.DefaultInstanceRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	nic, err := instanceAPI.WaitForMACAddress(&instance.WaitForMACAddressRequest{
		ServerID:      serverID,
		PrivateNicID:  privateNICID,
		Zone:          zone,
		Timeout:       scw.TimeDurationPtr(timeout),
		RetryInterval: scw.TimeDurationPtr(retryInterval),
	}, scw.WithContext(ctx))

	return nic, err
}

func waitForImage(ctx context.Context, api *instance.API, zone scw.Zone, id string, timeout time.Duration) (*instance.Image, error) {
	retryInterval := instancehelpers.DefaultInstanceRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	image, err := api.WaitForImage(&instance.WaitForImageRequest{
		ImageID:       id,
		Zone:          zone,
		Timeout:       scw.TimeDurationPtr(timeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))

	return image, err
}
