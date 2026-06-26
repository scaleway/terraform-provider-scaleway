package acctest_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestWaitForProjectIAM(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("succeeds after transient 403", func(t *testing.T) {
		t.Parallel()

		calls := 0

		err := acctest.WaitForProjectIAM(ctx, func(context.Context) error {
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

	t.Run("succeeds after permissions_denied", func(t *testing.T) {
		t.Parallel()

		calls := 0

		err := acctest.WaitForProjectIAM(ctx, func(context.Context) error {
			calls++
			if calls < 2 {
				return &scw.PermissionsDeniedError{}
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

	t.Run("returns non-403 error immediately", func(t *testing.T) {
		t.Parallel()

		calls := 0
		expected := errors.New("boom")

		err := acctest.WaitForProjectIAM(ctx, func(context.Context) error {
			calls++

			return expected
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if calls != 1 {
			t.Fatalf("expected 1 call, got %d", calls)
		}
	})
}
