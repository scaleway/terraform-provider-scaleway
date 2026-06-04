package mnqtestfuncs

import (
	"strings"
	"time"

	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

// NamespaceReadRetryTimeout is the retry window for MNQ namespace propagation in acceptance checks.
const NamespaceReadRetryTimeout = 60 * time.Second

// testAccMNQVCRProjectID is the default project ID in MNQ VCR cassettes.
const testAccMNQVCRProjectID = "1c29fc3f-d2dc-4267-8a05-ce9b616db4dc"

// ListProjectID returns project_id for MNQ acceptance tests: SDK default when set,
// otherwise the VCR placeholder so replay matches committed cassettes.
//
// Use this for resources that can coexist in the same project (e.g. NATS accounts).
// SNS/SQS namespace activation is one-per-project; those tests must keep a dedicated
// scaleway_account_project when running in parallel.
func ListProjectID(tt *acctest.TestTools) string {
	pid, ok := tt.Meta.ScwClient().GetDefaultProjectID()
	if ok {
		if s := strings.TrimSpace(pid); s != "" {
			return s
		}
	}

	return testAccMNQVCRProjectID
}
