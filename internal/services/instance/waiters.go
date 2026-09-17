package instance

import (
	"context"
	"time"

	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	instanceV2 "github.com/scaleway/scaleway-sdk-go/api/instance/v2alpha1"
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
		Timeout:       new(timeout),
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
		Timeout:       new(timeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))

	return server, err
}

func waitForPrivateNIC(ctx context.Context, instanceAPI *instanceV2.API, zone scw.Zone, privateNICID string, timeout time.Duration) (*instanceV2.PrivateNetworkInterface, error) {
	retryInterval := instancehelpers.DefaultInstanceRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	nic, err := instanceAPI.WaitForPrivateNetworkInterface(&instanceV2.WaitForPrivateNetworkInterfaceRequest{
		Zone:                      zone,
		PrivateNetworkInterfaceID: privateNICID,
		Timeout:                   new(timeout),
		RetryInterval:             new(retryInterval),
	}, scw.WithContext(ctx))

	return nic, err
}

func waitForMACAddress(ctx context.Context, instanceAPI *instance.API, zone scw.Zone, serverID string, privateNICID string, timeout time.Duration) error {
	retryInterval := instancehelpers.DefaultInstanceRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	_, err := instanceAPI.WaitForMACAddress(&instance.WaitForMACAddressRequest{
		ServerID:      serverID,
		PrivateNicID:  privateNICID,
		Zone:          zone,
		Timeout:       new(timeout),
		RetryInterval: new(retryInterval),
	}, scw.WithContext(ctx))

	return err
}

func waitForImage(ctx context.Context, api *instance.API, zone scw.Zone, id string, timeout time.Duration) (*instance.Image, error) {
	retryInterval := instancehelpers.DefaultInstanceRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	image, err := api.WaitForImage(&instance.WaitForImageRequest{
		ImageID:       id,
		Zone:          zone,
		Timeout:       new(timeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))

	return image, err
}

func waitForFilesystems(ctx context.Context, api *instance.API, zone scw.Zone, id string, timeout time.Duration) (*instance.Server, error) {
	retryInterval := instancehelpers.DefaultInstanceRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	server, err := api.WaitForServerFileSystem(&instance.WaitForServerFileSystemRequest{
		ServerID:      id,
		Zone:          zone,
		Timeout:       new(timeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))

	return server, err
}
