package kafka

import (
	"context"
	"time"

	kafka "github.com/scaleway/scaleway-sdk-go/api/kafka/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

func waitForKafkaCluster(ctx context.Context, api *kafka.API, region scw.Region, id string, timeout time.Duration) (*kafka.Cluster, error) {
	retryInterval := defaultWaitRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	req := &kafka.WaitForClusterRequest{
		ClusterID:     id,
		Region:        region,
		Timeout:       &timeout,
		RetryInterval: &retryInterval,
	}

	cluster, err := api.WaitForCluster(req, scw.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	return cluster, nil
}
