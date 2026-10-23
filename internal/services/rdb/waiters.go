package rdb

import (
	"context"
	"fmt"
	"time"

	"github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

func waitForRDBInstance(ctx context.Context, api *rdb.API, region scw.Region, id string, timeout time.Duration) (*rdb.Instance, error) {
	retryInterval := defaultWaitRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	var instance *rdb.Instance

	err := transport.RetryOnStatus(ctx, transport.IAMPropagation, func(ctx context.Context) error {
		var err error

		instance, err = api.WaitForInstance(&rdb.WaitForInstanceRequest{
			Region:        region,
			Timeout:       new(timeout),
			InstanceID:    id,
			RetryInterval: &retryInterval,
		}, scw.WithContext(ctx))
		if err != nil {
			return fmt.Errorf("waiting for rdb instance to be ready: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("waiting for rdb instance to be ready: %w", err)
	}

	return instance, nil
}

func waitForRDBDatabaseBackup(ctx context.Context, api *rdb.API, region scw.Region, id string, timeout time.Duration) (*rdb.DatabaseBackup, error) {
	retryInterval := defaultWaitRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	var backup *rdb.DatabaseBackup

	err := transport.RetryOnStatus(ctx, transport.IAMPropagation, func(ctx context.Context) error {
		var err error

		backup, err = api.WaitForDatabaseBackup(&rdb.WaitForDatabaseBackupRequest{
			Region:           region,
			Timeout:          new(timeout),
			DatabaseBackupID: id,
			RetryInterval:    &retryInterval,
		}, scw.WithContext(ctx))
		if err != nil {
			return fmt.Errorf("waiting for rdb database backup to be ready: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("waiting for rdb database backup to be ready: %w", err)
	}

	return backup, nil
}

func waitForRDBReadReplica(ctx context.Context, api *rdb.API, region scw.Region, id string, timeout time.Duration) (*rdb.ReadReplica, error) {
	retryInterval := defaultWaitRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	var replica *rdb.ReadReplica

	err := transport.RetryOnStatus(ctx, transport.IAMPropagation, func(ctx context.Context) error {
		var err error

		replica, err = api.WaitForReadReplica(&rdb.WaitForReadReplicaRequest{
			Region:        region,
			Timeout:       new(timeout),
			ReadReplicaID: id,
			RetryInterval: &retryInterval,
		}, scw.WithContext(ctx))
		if err != nil {
			return fmt.Errorf("waiting for rdb read replica to be ready: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("waiting for rdb read replica to be ready: %w", err)
	}

	return replica, nil
}

func waitForRDBSnapshot(ctx context.Context, api *rdb.API, region scw.Region, snapshotID string, timeout time.Duration) (*rdb.Snapshot, error) {
	retryInterval := defaultWaitRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	var snapshot *rdb.Snapshot

	err := transport.RetryOnStatus(ctx, transport.IAMPropagation, func(ctx context.Context) error {
		var err error

		snapshot, err = api.WaitForSnapshot(&rdb.WaitForSnapshotRequest{
			Region:        region,
			Timeout:       new(timeout),
			SnapshotID:    snapshotID,
			RetryInterval: &retryInterval,
		}, scw.WithContext(ctx))
		if err != nil {
			return fmt.Errorf("waiting for rdb snapshot to be ready: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("waiting for rdb snapshot to be ready: %w", err)
	}

	return snapshot, nil
}
