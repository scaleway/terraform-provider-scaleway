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

	volume, err := blockAPI.WaitForVolumeAndReferences(&block.WaitForVolumeAndReferencesRequest{
		Zone:          zone,
		VolumeID:      id,
		RetryInterval: &retryInterval,
		Timeout:       scw.TimeDurationPtr(timeout),
	}, scw.WithContext(ctx))

	return volume, err
}

func waitForBlockVolumeToBeAvailable(ctx context.Context, blockAPI *block.API, zone scw.Zone, id string, timeout time.Duration) (*block.Volume, error) {
	retryInterval := defaultBlockRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	terminalStatus := block.VolumeStatusAvailable
	volume, err := blockAPI.WaitForVolumeAndReferences(&block.WaitForVolumeAndReferencesRequest{
		Zone:          zone,
		VolumeID:      id,
		RetryInterval: &retryInterval,
		Timeout:       scw.TimeDurationPtr(timeout),

		VolumeTerminalStatus: &terminalStatus,
	}, scw.WithContext(ctx))

	return volume, err
}

func waitForBlockSnapshot(ctx context.Context, blockAPI *block.API, zone scw.Zone, id string, timeout time.Duration) (*block.Snapshot, error) {
	retryInterval := defaultBlockRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	snapshot, err := blockAPI.WaitForSnapshot(&block.WaitForSnapshotRequest{
		Zone:          zone,
		SnapshotID:    id,
		RetryInterval: &retryInterval,
		Timeout:       scw.TimeDurationPtr(timeout),
	}, scw.WithContext(ctx))

	return snapshot, err
}

func waitForBlockSnapshotToBeAvailable(ctx context.Context, blockAPI *block.API, zone scw.Zone, id string, timeout time.Duration) (*block.Snapshot, error) {
	retryInterval := defaultBlockRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	terminalStatus := block.SnapshotStatusAvailable
	snapshot, err := blockAPI.WaitForSnapshot(&block.WaitForSnapshotRequest{
		Zone:          zone,
		SnapshotID:    id,
		RetryInterval: &retryInterval,
		Timeout:       scw.TimeDurationPtr(timeout),

		TerminalStatus: &terminalStatus,
	}, scw.WithContext(ctx))

	return snapshot, err
}
