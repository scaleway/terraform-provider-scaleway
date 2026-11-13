package transport_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

func TestRetryOnStatusValue(t *testing.T) {
	t.Parallel()

	noWaitRetryInterval := 0 * time.Second
	defaultWaitRetryInterval := transport.DefaultWaitRetryInterval

	t.Cleanup(func() { transport.DefaultWaitRetryInterval = defaultWaitRetryInterval })
	// do not wait on retry in unit tests
	transport.DefaultWaitRetryInterval = &noWaitRetryInterval

	ctx := context.Background()

	t.Run("succeeds after transient 403", func(t *testing.T) {
		t.Parallel()

		calls := 0

		_, err := transport.RetryOnStatusValue(ctx, transport.IAMPropagation, func(context.Context) (struct{}, error) {
			calls++
			if calls < 3 {
				return struct{}{}, &scw.ResponseError{StatusCode: http.StatusForbidden}
			}

			return struct{}{}, nil
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if calls != 3 {
			t.Fatalf("expected 3 calls, got %d", calls)
		}
	})

	t.Run("succeeds after transient 404", func(t *testing.T) {
		t.Parallel()

		calls := 0

		value, err := transport.RetryOnStatusValue(ctx, transport.NotFound, func(context.Context) (string, error) {
			calls++
			if calls < 2 {
				return "", &scw.ResponseError{StatusCode: http.StatusNotFound}
			}

			return "ok", nil
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if value != "ok" {
			t.Fatalf("expected ok, got %q", value)
		}

		if calls != 2 {
			t.Fatalf("expected 2 calls, got %d", calls)
		}
	})

	t.Run("retries on multiple status codes", func(t *testing.T) {
		t.Parallel()

		calls := 0

		_, err := transport.RetryOnStatusValue(ctx, transport.RetryPolicy{
			StatusCodes: []int{http.StatusForbidden, http.StatusNotFound},
			Timeout:     time.Minute,
			WaitTime:    time.Second,
		}, func(context.Context) (struct{}, error) {
			calls++
			if calls < 3 {
				return struct{}{}, &scw.ResponseError{StatusCode: http.StatusNotFound}
			}

			return struct{}{}, nil
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if calls != 3 {
			t.Fatalf("expected 3 calls, got %d", calls)
		}
	})

	t.Run("returns non-matching error immediately", func(t *testing.T) {
		t.Parallel()

		calls := 0
		expected := errors.New("boom")

		_, err := transport.RetryOnStatusValue(ctx, transport.IAMPropagation, func(context.Context) (struct{}, error) {
			calls++

			return struct{}{}, expected
		})
		if !errors.Is(err, expected) {
			t.Fatalf("expected %v, got %v", expected, err)
		}

		if calls != 1 {
			t.Fatalf("expected 1 call, got %d", calls)
		}
	})

	t.Run("returns matching error after timeout", func(t *testing.T) {
		t.Parallel()

		permDenied := &scw.PermissionsDeniedError{}

		_, err := transport.RetryOnStatusValue(ctx, transport.RetryPolicy{
			StatusCodes: []int{http.StatusForbidden},
			Timeout:     10 * time.Millisecond,
			WaitTime:    time.Second,
		}, func(context.Context) (struct{}, error) {
			return struct{}{}, permDenied
		})
		if !errors.Is(err, permDenied) {
			t.Fatalf("expected %v, got %v", permDenied, err)
		}
	})

	t.Run("returns context error when cancelled", func(t *testing.T) {
		t.Parallel()

		cctx, cancel := context.WithCancel(ctx)
		cancel()

		_, err := transport.RetryOnStatusValue(cctx, transport.IAMPropagation, func(context.Context) (struct{}, error) {
			return struct{}{}, &scw.ResponseError{StatusCode: http.StatusForbidden}
		})
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context.Canceled, got %v", err)
		}
	})
}

func TestRetryOnStatus(t *testing.T) {
	t.Parallel()

	noWaitRetryInterval := 0 * time.Second
	defaultWaitRetryInterval := transport.DefaultWaitRetryInterval

	t.Cleanup(func() { transport.DefaultWaitRetryInterval = defaultWaitRetryInterval })
	// do not wait on retry in unit tests
	transport.DefaultWaitRetryInterval = &noWaitRetryInterval

	ctx := context.Background()

	t.Run("succeeds after transient 403", func(t *testing.T) {
		t.Parallel()

		calls := 0

		err := transport.RetryOnStatus(ctx, transport.IAMPropagation, func(context.Context) error {
			calls++
			if calls < 2 {
				return &scw.ResponseError{StatusCode: http.StatusForbidden}
			}

			return nil
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if calls != 2 {
			t.Fatalf("expected 2 calls, got %d", calls)
		}
	})

	t.Run("returns non-matching error immediately", func(t *testing.T) {
		t.Parallel()

		calls := 0
		expected := errors.New("boom")

		err := transport.RetryOnStatus(ctx, transport.NotFound, func(context.Context) error {
			calls++

			return expected
		})
		if !errors.Is(err, expected) {
			t.Fatalf("expected %v, got %v", expected, err)
		}

		if calls != 1 {
			t.Fatalf("expected 1 call, got %d", calls)
		}
	})
}
