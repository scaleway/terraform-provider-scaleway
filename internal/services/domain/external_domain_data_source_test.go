package domain_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceDomainExternal_Basic(t *testing.T) {
	if acctest.TestDomain == "" {
		t.Skip("Test skipped: SCW_TEST_DOMAIN must be set")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceDomainExternalConfigBasic(acctest.TestDomain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaleway_domain_external.test", "domain", acctest.TestDomain),
					resource.TestCheckResourceAttr("data.scaleway_domain_external.test", "status", "active"),
					resource.TestCheckResourceAttrSet("data.scaleway_domain_external.test", "id"),
					resource.TestCheckResourceAttrSet("data.scaleway_domain_external.test", "ns_servers.#"),
					resource.TestCheckResourceAttrSet("data.scaleway_domain_external.test", "project_id"),
				),
			},
		},
	})
}

func testAccDataSourceDomainExternalConfigBasic(domainName string) string {
	return fmt.Sprintf(`
data "scaleway_domain_external" "test" {
  domain = "%s"
}
`, domainName)
}
