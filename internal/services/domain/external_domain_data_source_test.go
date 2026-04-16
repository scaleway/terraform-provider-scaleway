package domain_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceDomainExternalDomain_Basic(t *testing.T) {
	if acctest.TestDomain == "" {
		t.Skip("Test skipped: SCW_TEST_DOMAIN must be set")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	subdomain := "tf-acc-ext-ds"
	domainName := fmt.Sprintf("%s.%s", subdomain, acctest.TestDomain)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCheckExternalDomainDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceDomainExternalDomainConfigBasic(domainName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_domain_external_domain.test", "domain", domainName),
					resource.TestCheckResourceAttrSet("scaleway_domain_external_domain.test", "validation_token"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_domain_external_domain.test", "domain",
						"scaleway_domain_external_domain.test", "domain",
					),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_domain_external_domain.test", "project_id",
						"scaleway_domain_external_domain.test", "project_id",
					),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_domain_external_domain.test", "organization_id",
						"scaleway_domain_external_domain.test", "organization_id",
					),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_domain_external_domain.test", "status",
						"scaleway_domain_external_domain.test", "status",
					),
				),
			},
		},
	})
}

func testAccDataSourceDomainExternalDomainConfigBasic(domainName string) string {
	return fmt.Sprintf(`
resource "scaleway_domain_external_domain" "test" {
  domain = "%s"
}

data "scaleway_domain_external_domain" "test" {
  domain     = scaleway_domain_external_domain.test.domain
  project_id = scaleway_domain_external_domain.test.project_id
}
`, domainName)
}
