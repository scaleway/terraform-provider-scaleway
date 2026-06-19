package rdb

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

var (
	maxRetriesOnForbidden = 3
	waitTime              = 1 * time.Second
)

// Mitigate transient permission issue during provisioning,
// caused by IAM permissions propagation delta.
func retryOn403(ctx context.Context, fn func() error) error {
	var lastErr error

	for range maxRetriesOnForbidden {
		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		var respErr *scw.ResponseError
		if errors.As(lastErr, &respErr) && respErr.StatusCode == http.StatusForbidden {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(waitTime):
				continue
			}
		}

		return lastErr
	}

	return lastErr
}

func waitForRDBInstance(ctx context.Context, api *rdb.API, region scw.Region, id string, timeout time.Duration) (*rdb.Instance, error) {
	retryInterval := defaultWaitRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	var instance *rdb.Instance

	err := retryOn403(ctx, func() error {
		var err error

		instance, err = api.WaitForInstance(&rdb.WaitForInstanceRequest{
			Region:        region,
			Timeout:       new(timeout),
			InstanceID:    id,
			RetryInterval: &retryInterval,
		}, scw.WithContext(ctx))

		return err
	})

	return instance, err
}

func waitForRDBDatabaseBackup(ctx context.Context, api *rdb.API, region scw.Region, id string, timeout time.Duration) (*rdb.DatabaseBackup, error) {
	retryInterval := defaultWaitRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	var backup *rdb.DatabaseBackup

	err := retryOn403(ctx, func() error {
		var err error

		backup, err = api.WaitForDatabaseBackup(&rdb.WaitForDatabaseBackupRequest{
			Region:           region,
			Timeout:          new(timeout),
			DatabaseBackupID: id,
			RetryInterval:    &retryInterval,
		}, scw.WithContext(ctx))

		return err
	})

	return backup, err
}

func waitForRDBReadReplica(ctx context.Context, api *rdb.API, region scw.Region, id string, timeout time.Duration) (*rdb.ReadReplica, error) {
	retryInterval := defaultWaitRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	var replica *rdb.ReadReplica

	err := retryOn403(ctx, func() error {
		var err error

		replica, err = api.WaitForReadReplica(&rdb.WaitForReadReplicaRequest{
			Region:        region,
			Timeout:       new(timeout),
			ReadReplicaID: id,
			RetryInterval: &retryInterval,
		}, scw.WithContext(ctx))

		return err
	})

	return replica, err
}

func waitForRDBSnapshot(ctx context.Context, api *rdb.API, region scw.Region, snapshotID string, timeout time.Duration) (*rdb.Snapshot, error) {
	retryInterval := defaultWaitRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	var snapshot *rdb.Snapshot

	err := retryOn403(ctx, func() error {
		var err error

		snapshot, err = api.WaitForSnapshot(&rdb.WaitForSnapshotRequest{
			Region:        region,
			Timeout:       new(timeout),
			SnapshotID:    snapshotID,
			RetryInterval: &retryInterval,
		}, scw.WithContext(ctx))

		return err
	})

	return snapshot, err
}
