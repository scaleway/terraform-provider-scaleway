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

func domainCheckPending(status tem.DomainStatus) bool {
	return status == tem.DomainStatusPending ||
		status == tem.DomainStatusUnchecked ||
		status == tem.DomainStatusAutoconfiguring
}

// WaitForDomainCheck polls CheckDomain until the domain leaves a pending status.
func WaitForDomainCheck(ctx context.Context, api *tem.API, region scw.Region, domainID string, timeout time.Duration) (*tem.Domain, error) {
	retryInterval := defaultDomainRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	deadline := time.Now().Add(timeout)

	for {
		domainCheck, err := api.CheckDomain(&tem.CheckDomainRequest{
			Region:   region,
			DomainID: domainID,
		}, scw.WithContext(ctx))
		if err == nil && domainCheck != nil && !domainCheckPending(domainCheck.Status) {
			return domainCheck, nil
		}

		remaining := time.Until(deadline)
		if remaining <= 0 {
			if err != nil {
				return nil, err
			}

			return nil, fmt.Errorf("domain validation did not complete in %d seconds", int(timeout.Seconds()))
		}

		wait := retryInterval
		if remaining < wait {
			wait = remaining
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(wait):
		}
	}
}
