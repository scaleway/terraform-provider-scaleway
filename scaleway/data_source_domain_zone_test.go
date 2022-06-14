package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayDataSourceDomainZone_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	testDNSZone := "test-zone2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayDomainZoneDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource scaleway_domain_zone main {
						domain    = "%s"
						subdomain = "%s"
					}

					data scaleway_domain_zone test {
						domain    = scaleway_domain_zone.main.domain
						subdomain = scaleway_domain_zone.main.subdomain
					}
				`, testDomain, testDNSZone),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayDomainZoneExists(tt, "data.scaleway_domain_zone.test"),
					resource.TestCheckResourceAttr("data.scaleway_domain_zone.test", "subdomain", testDNSZone),
					resource.TestCheckResourceAttr("data.scaleway_domain_zone.test", "domain", testDomain),
					resource.TestCheckResourceAttr("data.scaleway_domain_zone.test", "status", "active"),
				),
			},
		},
	})
}
