package scaleway_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccScalewayDataSourceLbIPs_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.TestAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayLbIPDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_lb_ip ip1 {
					}
				`,
			},
			{
				Config: `
					resource scaleway_lb_ip ip1 {
					}
					resource scaleway_lb_ip ip2 {
					}

					data "scaleway_lb_ips" "lbs_by_cidr_range" {
						ip_cidr_range = "0.0.0.0/0"
						depends_on = [scaleway_lb_ip.ip1, scaleway_lb_ip.ip2]
					}
					data "scaleway_lb_ips" "lbs_by_cidr_range_other_zone" {
						ip_cidr_range  = "0.0.0.0/0"
						zone = "fr-par-2"
						depends_on = [scaleway_lb_ip.ip1, scaleway_lb_ip.ip2]
					}
					`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scaleway_lb_ips.lbs_by_cidr_range", "ips.0.id"),
					resource.TestCheckResourceAttrSet("data.scaleway_lb_ips.lbs_by_cidr_range", "ips.0.ip_address"),
					resource.TestCheckResourceAttrSet("data.scaleway_lb_ips.lbs_by_cidr_range", "ips.1.id"),
					resource.TestCheckResourceAttrSet("data.scaleway_lb_ips.lbs_by_cidr_range", "ips.1.ip_address"),

					resource.TestCheckNoResourceAttr("data.scaleway_lb_ips.lbs_by_cidr_range_other_zone", "ips.0.id"),
				),
			},
		},
	})
}
