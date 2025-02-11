package inference

import (
	"context"
	"time"

	inference "github.com/scaleway/scaleway-sdk-go/api/inference/v1beta1"
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
		Timeout:       scw.TimeDurationPtr(timeout),
	}, scw.WithContext(ctx))

	return deployment, err
}
