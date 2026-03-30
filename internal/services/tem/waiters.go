package tem

import (
	"context"
	"fmt"
	"time"

	tem "github.com/scaleway/scaleway-sdk-go/api/tem/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

func WaitForDomain(ctx context.Context, api *tem.API, region scw.Region, id string, timeout time.Duration) (*tem.Domain, error) {
	retryInterval := defaultDomainRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	domain, err := api.WaitForDomain(&tem.WaitForDomainRequest{
		Region:        region,
		DomainID:      id,
		RetryInterval: &retryInterval,
		Timeout:       scw.TimeDurationPtr(timeout),
	}, scw.WithContext(ctx))

	return domain, err
}

func WaitForDomainAutoconfig(ctx context.Context, api *tem.API, region scw.Region, id string, want bool, timeout time.Duration) error {
	retryInterval := defaultDomainRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	ticker := time.NewTicker(retryInterval)
	defer ticker.Stop()

	deadline := time.Now().Add(timeout)
	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		domain, err := api.GetDomain(&tem.GetDomainRequest{
			Region:   region,
			DomainID: id,
		}, scw.WithContext(ctx))
		if err != nil {
			return err
		}

		if domain.Autoconfig == want {
			return nil
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for domain autoconfig=%v (last autoconfig=%v)", want, domain.Autoconfig)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
