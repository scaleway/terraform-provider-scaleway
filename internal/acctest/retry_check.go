package acctest

import (
	"context"

	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

// RetryCheckOn403 retries a read performed inside an acceptance Check function while IAM
// permissions are still propagating. Provider CRUD is guarded by transport.RetryOn403; Check
// functions call the API directly, so without the same guard a freshly created resource can fail
// its check on a transient 403 (IAM caches are per-instance and not monotonic). The retry is bounded
// by transport.IAMPropagationTimeout and only triggers on HTTP 403.
func RetryCheckOn403(fn func() error) error {
	return transport.RetryOn403(context.Background(), fn)
}
