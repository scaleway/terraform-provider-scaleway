package instance

import (
	"context"
	"fmt"
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
	if err != nil {
		return nil, fmt.Errorf("waiting for snapshot: %w", err)
	}

	return snapshot, nil
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
	if err != nil {
		return nil, fmt.Errorf("waiting for server: %w", err)
	}

	return server, nil
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
	if err != nil {
		return nil, fmt.Errorf("waiting for private network interface: %w", err)
	}

	return nic, nil
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
	if err != nil {
		return fmt.Errorf("waiting for MAC address: %w", err)
	}

	return nil
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
	if err != nil {
		return nil, fmt.Errorf("waiting for image: %w", err)
	}

	return image, nil
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
	if err != nil {
		return nil, fmt.Errorf("waiting for server filesystems: %w", err)
	}

	return server, nil
}
