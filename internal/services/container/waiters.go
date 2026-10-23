package container

import (
	"context"
	"fmt"
	"time"

	"github.com/scaleway/scaleway-sdk-go/api/container/v1"
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
	if err != nil {
		return nil, fmt.Errorf("waiting for namespace: %w", err)
	}

	return ns, nil
}

func waitForTrigger(ctx context.Context, api *container.API, triggerID string, region scw.Region, timeout time.Duration) (*container.Trigger, error) {
	retryInterval := defaultTriggerRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	request := container.WaitForTriggerRequest{
		TriggerID:     triggerID,
		Region:        region,
		Timeout:       new(timeout),
		RetryInterval: &retryInterval,
	}

	trigger, err := api.WaitForTrigger(&request, scw.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("waiting for trigger: %w", err)
	}

	return trigger, nil
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

	container, err := api.WaitForContainer(&request, scw.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("waiting for container: %w", err)
	}

	return container, nil
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

	domain, err := api.WaitForDomain(&request, scw.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("waiting for domain: %w", err)
	}

	return domain, nil
}
