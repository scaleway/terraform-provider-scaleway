package messageq

import (
	"context"
	"fmt"
	"time"

	messageqapi "github.com/scaleway/scaleway-sdk-go/api/messageq/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

func waitForDeployment(
	ctx context.Context,
	api *messageqapi.API,
	region scw.Region,
	deploymentID string,
	timeout time.Duration,
) (*messageqapi.Deployment, error) {
	retryInterval := defaultWaitRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	var deployment *messageqapi.Deployment

	err := transport.RetryOn403(ctx, func() error {
		var err error

		deployment, err = api.WaitForDeployment(&messageqapi.WaitForDeploymentRequest{
			Region:        region,
			DeploymentID:  deploymentID,
			Timeout:       &timeout,
			RetryInterval: &retryInterval,
		}, scw.WithContext(ctx))

		return err
	})

	return deployment, err
}

// waitForEndpointsDeleted polls GetDeployment until every endpoint ID in endpointIDs
// is no longer present in the deployment's endpoint list.
func waitForEndpointsDeleted(
	ctx context.Context,
	api *messageqapi.API,
	region scw.Region,
	deploymentID string,
	endpointIDs []string,
	timeout time.Duration,
) error {
	if len(endpointIDs) == 0 {
		return nil
	}

	retryInterval := defaultWaitRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	idSet := make(map[string]bool, len(endpointIDs))
	for _, id := range endpointIDs {
		idSet[id] = true
	}

	start := time.Now()

	for {
		var deployment *messageqapi.Deployment

		err := transport.RetryOn403(ctx, func() error {
			var err error

			deployment, err = api.GetDeployment(&messageqapi.GetDeploymentRequest{
				Region:       region,
				DeploymentID: deploymentID,
			}, scw.WithContext(ctx))

			return err
		})
		if err != nil {
			return err
		}

		anyStillPresent := false

		for _, ep := range deployment.Endpoints {
			if ep == nil {
				continue
			}

			if idSet[ep.ID] {
				anyStillPresent = true

				break
			}
		}

		if !anyStillPresent {
			return nil
		}

		if time.Since(start) > timeout {
			return fmt.Errorf("timeout waiting for endpoints %v to be deleted from deployment %s", endpointIDs, deploymentID)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(retryInterval):
		}
	}
}
