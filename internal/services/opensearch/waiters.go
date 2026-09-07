package opensearch

import (
	"context"
	"time"

	searchdbapi "github.com/scaleway/scaleway-sdk-go/api/searchdb/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

func waitForDeployment(
	ctx context.Context,
	api *searchdbapi.API,
	region scw.Region,
	deploymentID string,
	timeout time.Duration,
) (*searchdbapi.Deployment, error) {
	retryInterval := defaultWaitRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	var deployment *searchdbapi.Deployment

	err := transport.RetryOn403(ctx, func() error {
		var err error

		deployment, err = api.WaitForDeployment(&searchdbapi.WaitForDeploymentRequest{
			Region:        region,
			DeploymentID:  deploymentID,
			Timeout:       &timeout,
			RetryInterval: &retryInterval,
		}, scw.WithContext(ctx))

		return err
	})

	return deployment, err
}
