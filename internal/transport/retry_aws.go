package transport

import (
	"context"
	"errors"
	"time"

	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
)

// RetryWhenAWSErrCodeEquals retries a function when it returns a specific AWS error
func RetryWhenAWSErrCodeEquals[T any](ctx context.Context, codes []string, config *RetryWhenConfig[T]) (T, error) { //nolint: ireturn
	return retryWhen(ctx, config, func(err error) bool {
		return tfawserr.ErrCodeEquals(err, codes...)
	})
}

// RetryWhenAWSErrCodeNotEquals retries a function until it returns a specific AWS error
func RetryWhenAWSErrCodeNotEquals[T any](ctx context.Context, codes []string, config *RetryWhenConfig[T]) (T, error) { //nolint: ireturn
	return retryWhen(ctx, config, func(err error) bool {
		if err == nil {
			return true
		}

		return !tfawserr.ErrCodeEquals(err, codes...)
	})
}

// retryWhen executes the function passed in the configuration object until the timeout is reached or the context is cancelled.
// It will retry if the shouldRetry function returns true. It will stop if the shouldRetry function returns false.
func retryWhen[T any](ctx context.Context, config *RetryWhenConfig[T], shouldRetry func(error) bool) (T, error) { //nolint: ireturn
	retryInterval := config.Interval
	if DefaultWaitRetryInterval != nil {
		retryInterval = *DefaultWaitRetryInterval
	}

	timer := time.NewTimer(config.Timeout)

	for {
		result, err := config.Function()
		if shouldRetry(err) {
			select {
			case <-timer.C:
				return result, ErrRetryWhenTimeout
			case <-ctx.Done():
				return result, ctx.Err()
			default:
				time.Sleep(retryInterval) // lintignore:R018
				continue
			}
		}

		return result, err
	}
}

type RetryWhenConfig[T any] struct {
	Timeout  time.Duration
	Interval time.Duration
	Function func() (T, error)
}

var ErrRetryWhenTimeout = errors.New("timeout reached")
