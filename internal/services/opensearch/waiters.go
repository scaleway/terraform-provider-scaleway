package opensearch

import (
	"context"
	"time"

	searchdbapi "github.com/scaleway/scaleway-sdk-go/api/searchdb/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func waitForDeployment(
	ctx context.Context,
	api *searchdbapi.API,
	region scw.Region,
	deploymentID string,
	timeout time.Duration,
) (*searchdbapi.Deployment, error) {
	retryInterval := defaultWaitRetryInterval

	return api.WaitForDeployment(&searchdbapi.WaitForDeploymentRequest{
		Region:        region,
		DeploymentID:  deploymentID,
		Timeout:       &timeout,
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))
}
