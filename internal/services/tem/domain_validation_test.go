package tem_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

// NOTE: TestAccDomainValidation_Validation was removed as it's a duplicate of TestAccDomain_Autoconfig in domain_test.go
// Both tests verify autoconfig + validation, so we keep only one to avoid CI timeout issues

func TestAccDomainValidation_TimeoutError(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	domainName := "terraform-timeout.test.local"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isDomainDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
                    resource scaleway_tem_domain cr01 {
                        name       = "%s"
                        accept_tos = true
                        autoconfig = false
                    }

                    resource scaleway_tem_domain_validation valid {
                        domain_id = scaleway_tem_domain.cr01.id
                        region    = scaleway_tem_domain.cr01.region
                        timeout   = 1
                    }
                `, domainName),
				ExpectError: regexp.MustCompile("(?i)domain validation did not complete"),
			},
		},
	})
}
