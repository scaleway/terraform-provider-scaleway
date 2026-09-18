package mailboxtestfuncs

import (
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

// CassetteProjectID is the project ID embedded in mailbox VCR cassettes.
const CassetteProjectID = "46fd79d8-1a35-4548-bfb8-03df51a0ebae"

// TestProjectID returns the project ID to use in acceptance tests.
// Replay always uses CassetteProjectID so Terraform config matches cassette
// responses and tests work without SCW_DEFAULT_PROJECT_ID. Recording uses the
// client's default project.
func TestProjectID(tt *acctest.TestTools) string {
	tt.T.Helper()

	if !*acctest.UpdateCassettes {
		return CassetteProjectID
	}

	projectID, ok := tt.Meta.ScwClient().GetDefaultProjectID()
	if !ok {
		tt.T.Fatal("default project ID is required to record mailbox cassettes")
	}

	return projectID
}
