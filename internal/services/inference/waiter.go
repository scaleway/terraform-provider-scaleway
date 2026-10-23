package inference

import (
	"context"
	"fmt"
	"time"

	"github.com/scaleway/scaleway-sdk-go/api/inference/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

func waitForDeployment(ctx context.Context, inferenceAPI *inference.API, region scw.Region, id string, timeout time.Duration) (*inference.Deployment, error) {
	retryInterval := defaultDeploymentRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	deployment, err := inferenceAPI.WaitForDeployment(&inference.WaitForDeploymentRequest{
		Region:        region,
		DeploymentID:  id,
		RetryInterval: &retryInterval,
		Timeout:       new(timeout),
	}, scw.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("waiting for deployment: %w", err)
	}

	return deployment, nil
}

func waitForModel(ctx context.Context, inferenceAPI *inference.API, region scw.Region, id string, timeout time.Duration) (*inference.Model, error) {
	retryInterval := defaultModelRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	model, err := inferenceAPI.WaitForModel(&inference.WaitForModelRequest{
		ModelID:       id,
		Region:        region,
		RetryInterval: &retryInterval,
		Timeout:       new(timeout),
	}, scw.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("waiting for model: %w", err)
	}

	return model, nil
}
