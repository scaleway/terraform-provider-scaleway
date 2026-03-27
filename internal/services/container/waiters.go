package container

import (
	"context"
	"time"

	container "github.com/scaleway/scaleway-sdk-go/api/container/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

func waitForNamespace(ctx context.Context, containerAPI *container.API, region scw.Region, namespaceID string, timeout time.Duration) (*container.Namespace, error) {
	retryInterval := DefaultContainerRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	ns, err := containerAPI.WaitForNamespace(&container.WaitForNamespaceRequest{
		Region:        region,
		NamespaceID:   namespaceID,
		RetryInterval: &retryInterval,
		Timeout:       new(timeout),
	}, scw.WithContext(ctx))

	return ns, err
}

func waitForCron(ctx context.Context, api *container.API, cronID string, region scw.Region, timeout time.Duration) (*container.Cron, error) {
	retryInterval := DefaultContainerRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	request := container.WaitForCronRequest{
		CronID:        cronID,
		Region:        region,
		Timeout:       new(timeout),
		RetryInterval: &retryInterval,
	}

	return api.WaitForCron(&request, scw.WithContext(ctx))
}

func waitForContainer(ctx context.Context, api *container.API, containerID string, region scw.Region, timeout time.Duration) (*container.Container, error) {
	retryInterval := DefaultContainerRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	request := container.WaitForContainerRequest{
		ContainerID:   containerID,
		Region:        region,
		Timeout:       new(timeout),
		RetryInterval: &retryInterval,
	}

	return api.WaitForContainer(&request, scw.WithContext(ctx))
}

func waitForDomain(ctx context.Context, api *container.API, domainID string, region scw.Region, timeout time.Duration) (*container.Domain, error) {
	retryInterval := DefaultContainerRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	request := container.WaitForDomainRequest{
		DomainID:      domainID,
		Region:        region,
		Timeout:       new(timeout),
		RetryInterval: &retryInterval,
	}

	return api.WaitForDomain(&request, scw.WithContext(ctx))
}
