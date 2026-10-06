package transport_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

func TestRetryOn403(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("succeeds after transient 403", func(t *testing.T) {
		t.Parallel()

		calls := 0

		err := transport.RetryOn403(ctx, func() error {
			calls++
			if calls < 3 {
				return &scw.ResponseError{StatusCode: http.StatusForbidden}
			}

			return nil
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if calls != 3 {
			t.Fatalf("expected 3 calls, got %d", calls)
		}
	})

	t.Run("returns non-403 error immediately", func(t *testing.T) {
		t.Parallel()

		calls := 0
		expected := errors.New("boom")

		err := transport.RetryOn403(ctx, func() error {
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

func TestRetryOn403Value(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	calls := 0

	value, err := transport.RetryOn403Value(ctx, func() (string, error) {
		calls++
		if calls < 2 {
			return "", &scw.ResponseError{StatusCode: http.StatusForbidden}
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
}
