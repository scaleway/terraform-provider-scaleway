package datawarehouse

import (
	"context"
	"time"

	datawarehouse "github.com/scaleway/scaleway-sdk-go/api/datawarehouse/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

func waitForDatawarehouseDeployment(ctx context.Context, api *datawarehouse.API, region scw.Region, id string, timeout time.Duration) (*datawarehouse.Deployment, error) {
	retryInterval := defaultWaitRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	req := &datawarehouse.WaitForDeploymentRequest{
		DeploymentID:  id,
		Region:        region,
		Timeout:       &timeout,
		RetryInterval: &retryInterval,
	}

	deployment, err := api.WaitForDeployment(req, scw.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	return deployment, nil
}
