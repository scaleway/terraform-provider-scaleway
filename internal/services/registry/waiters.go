package registry

import (
	"context"
	"fmt"
	"time"

	"github.com/scaleway/scaleway-sdk-go/api/registry/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

func WaitForNamespace(ctx context.Context, api *registry.API, region scw.Region, id string, timeout time.Duration) (*registry.Namespace, error) {
	retryInterval := defaultNamespaceRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	ns, err := api.WaitForNamespace(&registry.WaitForNamespaceRequest{
		Region:        region,
		NamespaceID:   id,
		RetryInterval: &retryInterval,
		Timeout:       scw.TimeDurationPtr(timeout),
	}, scw.WithContext(ctx))

	return ns, err
}

func waitForNamespaceDelete(ctx context.Context, api *registry.API, region scw.Region, id string, timeout time.Duration) (*registry.Namespace, error) {
	retryInterval := defaultNamespaceRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	terminalStatus := map[registry.NamespaceStatus]struct{}{
		registry.NamespaceStatusReady:    {},
		registry.NamespaceStatusLocked:   {},
		registry.NamespaceStatusError:    {},
		registry.NamespaceStatusUnknown:  {},
		registry.NamespaceStatusDeleting: {},
	}

	start := time.Now()
	for {
		ns, err := api.GetNamespace(&registry.GetNamespaceRequest{
			Region:      region,
			NamespaceID: id,
		}, scw.WithContext(ctx))
		if err != nil {
			return nil, err
		}

		if _, ok := terminalStatus[ns.Status]; ok {
			return ns, nil
		}

		if time.Since(start) > timeout {
			return nil, fmt.Errorf("timeout while waiting for namespace %s to be deleted", id)
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(retryInterval):
		}
	}
}
