package mnq

import (
	"context"
	"strings"
	"time"

	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

func retryMNQNamespaceRead(ctx context.Context, action func() error) error {
	_, err := RetryMNQNamespaceReadValue(ctx, func() (struct{}, error) {
		return struct{}{}, action()
	})

	return err
}

// RetryMNQNamespaceReadValue retries fn on transient MNQ namespace errors (IAM 403 propagation, 404).
func RetryMNQNamespaceReadValue[T any](ctx context.Context, fn func() (T, error)) (T, error) {
	var (
		result  T
		lastErr error
	)

	wait := transport.RetryOn403WaitTime
	if transport.DefaultWaitRetryInterval != nil {
		wait = *transport.DefaultWaitRetryInterval
	}

	deadline := time.Now().Add(transport.IAMPropagationTimeout)

	for {
		result, lastErr = fn()
		if lastErr == nil {
			return result, nil
		}

		if !isMNQNamespaceReadRetryableError(lastErr) || time.Now().After(deadline) {
			return result, lastErr
		}

		select {
		case <-ctx.Done():
			return result, ctx.Err()
		case <-time.After(wait):
		}
	}
}

func isMNQNamespaceReadRetryableError(err error) bool {
	if err == nil {
		return false
	}

	if httperrors.Is404(err) && strings.Contains(err.Error(), "resource namespace") {
		return true
	}

	return httperrors.Is403(err)
}
