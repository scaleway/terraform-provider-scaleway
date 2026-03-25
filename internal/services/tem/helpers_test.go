package tem_test

import (
	"os"
	"strings"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

// scwAcceptanceTestFakeSecretKey matches placeholder SCW_SECRET_KEY in .github/workflows/acceptance-tests.yaml (VCR replay).
const scwAcceptanceTestFakeSecretKey = "11111111-1111-1111-1111-111111111111"

// testAccDomainZoneVCRProjectID matches anonymized Scaleway DNS zone VCR cassettes (CI replay with placeholder credentials).
const testAccDomainZoneVCRProjectID = "105bdce1-64c0-48ab-899d-868455867ecf"

// testAccDomainZoneProjectID is the project_id for scaleway_domain_zone in acceptance configs: SDK default when set, else VCR placeholder.
func testAccDomainZoneProjectID(tt *acctest.TestTools) string {
	pid, ok := tt.Meta.ScwClient().GetDefaultProjectID()
	if ok {
		if s := strings.TrimSpace(pid); s != "" {
			return s
		}
	}

	return testAccDomainZoneVCRProjectID
}

// testAccWebhookDomainZoneSubdomain is stable for CI VCR (fake key or -cassettes) and random for live runs to avoid duplicate DNS zones.
func testAccWebhookDomainZoneSubdomain() string {
	if *acctest.UpdateCassettes {
		return "tf-webhook-acc"
	}
	if os.Getenv("SCW_SECRET_KEY") == scwAcceptanceTestFakeSecretKey {
		return "tf-webhook-acc"
	}

	return sdkacctest.RandomWithPrefix("tf-webhook")
}
