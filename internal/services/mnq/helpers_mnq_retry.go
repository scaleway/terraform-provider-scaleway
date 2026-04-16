package mnq

import (
	"context"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
)

const mnqNamespaceReadRetryTimeout = 120 * time.Second

func retryMNQNamespaceRead(ctx context.Context, action func() error) error {
	return retry.RetryContext(ctx, mnqNamespaceReadRetryTimeout, func() *retry.RetryError {
		err := action()
		if err == nil {
			return nil
		}

		if isMNQNamespaceReadRetryableError(err) {
			return retry.RetryableError(err)
		}

		return retry.NonRetryableError(err)
	})
}

func isMNQNamespaceReadRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Namespace propagation can briefly return 404 on fresh resources.
	if httperrors.Is404(err) && strings.Contains(err.Error(), "resource namespace") {
		return true
	}

	// Some MNQ read calls intermittently return permission denied on just-created resources.
	return strings.Contains(err.Error(), "insufficient permissions: read namespace")
}
