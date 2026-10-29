package function

import (
	"context"
	"time"

	function "github.com/scaleway/scaleway-sdk-go/api/function/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

func waitForNamespace(ctx context.Context, functionAPI *function.API, region scw.Region, id string, timeout time.Duration) (*function.Namespace, error) {
	retryInterval := DefaultFunctionRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	ns, err := functionAPI.WaitForNamespace(&function.WaitForNamespaceRequest{
		Region:        region,
		NamespaceID:   id,
		RetryInterval: &retryInterval,
		Timeout:       new(timeout),
	}, scw.WithContext(ctx))

	return ns, err
}

func waitForFunction(ctx context.Context, functionAPI *function.API, region scw.Region, id string, timeout time.Duration) (*function.Function, error) {
	retryInterval := DefaultFunctionRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	f, err := functionAPI.WaitForFunction(&function.WaitForFunctionRequest{
		Region:        region,
		FunctionID:    id,
		RetryInterval: &retryInterval,
		Timeout:       new(timeout),
	}, scw.WithContext(ctx))

	return f, err
}

func waitForCron(ctx context.Context, functionAPI *function.API, region scw.Region, cronID string, timeout time.Duration) (*function.Cron, error) {
	retryInterval := DefaultFunctionRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	return functionAPI.WaitForCron(&function.WaitForCronRequest{
		Region:        region,
		CronID:        cronID,
		RetryInterval: &retryInterval,
		Timeout:       new(timeout),
	}, scw.WithContext(ctx))
}

func waitForDomain(ctx context.Context, functionAPI *function.API, region scw.Region, id string, timeout time.Duration) (*function.Domain, error) {
	retryInterval := DefaultFunctionRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	domain, err := functionAPI.WaitForDomain(&function.WaitForDomainRequest{
		Region:        region,
		DomainID:      id,
		RetryInterval: &retryInterval,
		Timeout:       new(timeout),
	}, scw.WithContext(ctx))

	return domain, err
}

func waitForTrigger(ctx context.Context, functionAPI *function.API, region scw.Region, id string, timeout time.Duration) (*function.Trigger, error) {
	retryInterval := DefaultFunctionRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	trigger, err := functionAPI.WaitForTrigger(&function.WaitForTriggerRequest{
		Region:        region,
		TriggerID:     id,
		RetryInterval: &retryInterval,
		Timeout:       new(timeout),
	}, scw.WithContext(ctx))

	return trigger, err
}
