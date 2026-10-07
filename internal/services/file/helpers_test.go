package file_test

import (
	"testing"

	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

// skipUnlessDefaultProjectID skips when SCW_DEFAULT_PROJECT_ID is unset.
// Framework Create resolves project_id via ExtractFrameworkProjectID and hard-fails
// without a client default; the OpenTofu ACC job does not inject that secret
// (unlike the Terraform job).
func skipUnlessDefaultProjectID(t *testing.T, tt *acctest.TestTools) {
	t.Helper()

	if _, ok := tt.Meta.ScwClient().GetDefaultProjectID(); !ok {
		t.Skip("No default project ID found, skipping test")
	}
}
