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

	return api.WaitForInstance(&rdb.WaitForInstanceRequest{
		Region:        region,
		Timeout:       new(timeout),
		InstanceID:    id,
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))
}

func waitForRDBDatabaseBackup(ctx context.Context, api *rdb.API, region scw.Region, id string, timeout time.Duration) (*rdb.DatabaseBackup, error) {
	retryInterval := defaultWaitRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	return api.WaitForDatabaseBackup(&rdb.WaitForDatabaseBackupRequest{
		Region:           region,
		Timeout:          new(timeout),
		DatabaseBackupID: id,
		RetryInterval:    &retryInterval,
	}, scw.WithContext(ctx))
}

func waitForRDBReadReplica(ctx context.Context, api *rdb.API, region scw.Region, id string, timeout time.Duration) (*rdb.ReadReplica, error) {
	retryInterval := defaultWaitRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	return api.WaitForReadReplica(&rdb.WaitForReadReplicaRequest{
		Region:        region,
		Timeout:       new(timeout),
		ReadReplicaID: id,
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))
}

func waitForRDBSnapshot(ctx context.Context, api *rdb.API, region scw.Region, snapshotID string, timeout time.Duration) (*rdb.Snapshot, error) {
	retryInterval := defaultWaitRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	return api.WaitForSnapshot(&rdb.WaitForSnapshotRequest{
		Region:        region,
		Timeout:       new(timeout),
		SnapshotID:    snapshotID,
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))
}

// waitForRDBInstanceEndpoints waits until expected endpoint types are present on the instance.
// This avoids persisting null computed attributes when endpoints are still being provisioned.
func waitForRDBInstanceEndpoints(
	ctx context.Context,
	api *rdb.API,
	region scw.Region,
	instanceID string,
	timeout time.Duration,
	wantPrivateNetwork, wantLoadBalancer bool,
) error {
	if !wantPrivateNetwork && !wantLoadBalancer {
		return nil
	}

	retryInterval := defaultWaitRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	if retryInterval <= 0 {
		retryInterval = time.Millisecond
	}

	ticker := time.NewTicker(retryInterval)
	defer ticker.Stop()

	deadline := time.Now().Add(timeout)

	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		res, err := waitForRDBInstance(ctx, api, region, instanceID, timeout)
		if err != nil {
			return err
		}

		if instanceEndpointsReady(res.Endpoints, wantPrivateNetwork, wantLoadBalancer) {
			return nil
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for RDB instance %q endpoints to be ready", instanceID)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func instanceEndpointsReady(endpoints []*rdb.Endpoint, wantPrivateNetwork, wantLoadBalancer bool) bool {
	hasPrivateNetwork := !wantPrivateNetwork
	hasLoadBalancer := !wantLoadBalancer

	for _, endpoint := range endpoints {
		if wantPrivateNetwork && endpoint.PrivateNetwork != nil && endpoint.ID != "" {
			hasPrivateNetwork = true
		}

		if wantLoadBalancer && endpoint.LoadBalancer != nil && endpoint.ID != "" {
			hasLoadBalancer = true
		}
	}

	return hasPrivateNetwork && hasLoadBalancer
}
