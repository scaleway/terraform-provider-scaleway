package k8s

import (
	"context"
	"fmt"
	"time"

	"github.com/scaleway/scaleway-sdk-go/api/k8s/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

func waitCluster(ctx context.Context, k8sAPI *k8s.API, region scw.Region, clusterID string, timeout time.Duration) (*k8s.Cluster, error) {
	retryInterval := defaultK8SRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	cluster, err := k8sAPI.WaitForCluster(&k8s.WaitForClusterRequest{
		ClusterID:     clusterID,
		Region:        region,
		Timeout:       new(timeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))

	return cluster, err
}

func waitClusterPool(ctx context.Context, k8sAPI *k8s.API, region scw.Region, clusterID string, timeout time.Duration) (*k8s.Cluster, error) {
	retryInterval := defaultK8SRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	return k8sAPI.WaitForClusterPool(&k8s.WaitForClusterRequest{
		ClusterID:     clusterID,
		Region:        region,
		Timeout:       new(timeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))
}

func waitClusterStatus(ctx context.Context, k8sAPI *k8s.API, cluster *k8s.Cluster, status k8s.ClusterStatus, timeout time.Duration) (*k8s.Cluster, error) {
	retryInterval := defaultK8SRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	cluster, err := k8sAPI.WaitForCluster(&k8s.WaitForClusterRequest{
		ClusterID:     cluster.ID,
		Region:        cluster.Region,
		Timeout:       new(timeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))
	if err != nil {
		if status == k8s.ClusterStatusDeleted && httperrors.Is404(err) {
			return cluster, nil
		}

		return cluster, err
	}

	return cluster, nil
}

func waitPoolReady(ctx context.Context, k8sAPI *k8s.API, region scw.Region, poolID string, timeout time.Duration) (*k8s.Pool, error) {
	retryInterval := defaultK8SRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	pool, err := k8sAPI.WaitForPool(&k8s.WaitForPoolRequest{
		PoolID:        poolID,
		Region:        region,
		Timeout:       new(timeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	if pool.Status != k8s.PoolStatusReady {
		return nil, fmt.Errorf("pool %s has state %s, wants %s", poolID, pool.Status, k8s.PoolStatusReady)
	}

	return pool, nil
}
