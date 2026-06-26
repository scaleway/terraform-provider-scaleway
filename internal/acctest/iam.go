package acctest

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
)

// iamPropagationBackoff waits up to ~62s to outlive IAM's ~60s permission cache.
var iamPropagationBackoff = []time.Duration{
	2 * time.Second,
	4 * time.Second,
	8 * time.Second,
	16 * time.Second,
	32 * time.Second,
}

// WaitForProjectIAM retries probe until it succeeds or a non-403 error is returned.
// probe should perform the same API call that fails transiently while IAM permissions propagate.
func WaitForProjectIAM(ctx context.Context, probe func(context.Context) error) error {
	var lastErr error

	for i, wait := range iamPropagationBackoff {
		lastErr = probe(ctx)
		if lastErr == nil {
			return nil
		}

		if !httperrors.Is403(lastErr) {
			return fmt.Errorf("waiting for IAM permissions: %w", lastErr)
		}

		if i == len(iamPropagationBackoff)-1 {
			break
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}
	}

	return fmt.Errorf("IAM permissions not propagated after retries: %w", lastErr)
}

// StoreResourceID saves a resource ID from state into id for use in later test steps.
func StoreResourceID(resourceName string, id *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		*id = rs.Primary.ID

		return nil
	}
}
