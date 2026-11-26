package block

import (
	"context"
	"time"

	"github.com/scaleway/scaleway-sdk-go/api/block/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

func waitForBlockVolume(ctx context.Context, blockAPI *block.API, zone scw.Zone, id string, timeout time.Duration) (*block.Volume, error) {
	retryInterval := defaultBlockRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	var volume *block.Volume

	err := transport.RetryOnStatus(ctx, transport.IAMPropagation, func(ctx context.Context) error {
		var err error

		volume, err = blockAPI.WaitForVolumeAndReferences(&block.WaitForVolumeAndReferencesRequest{
			Zone:          zone,
			VolumeID:      id,
			RetryInterval: &retryInterval,
			Timeout:       new(timeout),
		}, scw.WithContext(ctx))

		return err
	})

	return volume, err
}

func waitForBlockVolumeToBeAvailable(ctx context.Context, blockAPI *block.API, zone scw.Zone, id string, timeout time.Duration) (*block.Volume, error) {
	retryInterval := defaultBlockRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	var volume *block.Volume

	err := transport.RetryOnStatus(ctx, transport.IAMPropagation, func(ctx context.Context) error {
		var err error

		volume, err = blockAPI.WaitForVolumeAndReferences(&block.WaitForVolumeAndReferencesRequest{
			Zone:          zone,
			VolumeID:      id,
			RetryInterval: &retryInterval,
			Timeout:       new(timeout),

			VolumeTerminalStatus: new(block.VolumeStatusAvailable),
		}, scw.WithContext(ctx))

		return err
	})

	return volume, err
}

func waitForBlockSnapshot(ctx context.Context, blockAPI *block.API, zone scw.Zone, id string, timeout time.Duration) (*block.Snapshot, error) {
	retryInterval := defaultBlockRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	var snapshot *block.Snapshot

	err := transport.RetryOnStatus(ctx, transport.IAMPropagation, func(ctx context.Context) error {
		var err error

		snapshot, err = blockAPI.WaitForSnapshot(&block.WaitForSnapshotRequest{
			Zone:          zone,
			SnapshotID:    id,
			RetryInterval: &retryInterval,
			Timeout:       new(timeout),
		}, scw.WithContext(ctx))

		return err
	})

	return snapshot, err
}

func waitForBlockSnapshotToBeAvailable(ctx context.Context, blockAPI *block.API, zone scw.Zone, id string, timeout time.Duration) (*block.Snapshot, error) {
	retryInterval := defaultBlockRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	var snapshot *block.Snapshot

	err := transport.RetryOnStatus(ctx, transport.IAMPropagation, func(ctx context.Context) error {
		var err error

		snapshot, err = blockAPI.WaitForSnapshot(&block.WaitForSnapshotRequest{
			Zone:          zone,
			SnapshotID:    id,
			RetryInterval: &retryInterval,
			Timeout:       new(timeout),

			TerminalStatus: new(block.SnapshotStatusAvailable),
		}, scw.WithContext(ctx))

		return err
	})

	return snapshot, err
}
