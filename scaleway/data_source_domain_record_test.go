package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayDataSourceDomainRecord_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayDomainRecordDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource scaleway_domain_record main {
						dns_zone = "test-data-source.%s"
						name     = "www"
						type     = "A"
						data     = "1.2.3.4"
						ttl      = 3600
					}

					data scaleway_domain_record test {
						dns_zone  = "${scaleway_domain_record.main.dns_zone}"
						record_id = "${scaleway_domain_record.main.id}"
					}

					data scaleway_domain_record test2 {
						dns_zone = "${scaleway_domain_record.main.dns_zone}"
						name     = "${scaleway_domain_record.main.name}"
						type     = "${scaleway_domain_record.main.type}"
						data     = "${scaleway_domain_record.main.data}"
					}
				`, testDomain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayDomainRecordExists(tt, "data.scaleway_domain_record.test"),
					testAccCheckScalewayDomainRecordExists(tt, "data.scaleway_domain_record.test2"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_domain_record.test", "dns_zone",
						"scaleway_domain_record.main", "dns_zone"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_domain_record.test2", "id",
						"scaleway_domain_record.main", "id"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_domain_record.test2", "dns_zone",
						"scaleway_domain_record.main", "dns_zone"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_domain_record.test2", "name",
						"scaleway_domain_record.main", "name"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_domain_record.test2", "type",
						"scaleway_domain_record.main", "type"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_domain_record.test2", "data",
						"scaleway_domain_record.main", "data"),
				),
			},
		},
	})
}
