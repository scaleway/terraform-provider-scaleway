package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayDataSourceLbIPs_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
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

					data "scaleway_lb_ips" "lbs_by_ip_address" {
						ip_address = "${scaleway_lb_ip.ip1.ip_address}"
						depends_on = [scaleway_lb_ip.ip1, scaleway_lb_ip.ip2]
					}
					data "scaleway_lb_ips" "lbs_by_ip_address_other_zone" {
						ip_address  = "${scaleway_lb_ip.ip1.ip_address}"
						zone = "fr-par-2"
						depends_on = [scaleway_lb_ip.ip1, scaleway_lb_ip.ip2]
					}
					`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scaleway_lb_ips.lbs_by_ip_address", "ips.0.id"),
					resource.TestCheckResourceAttrSet("data.scaleway_lb_ips.lbs_by_ip_address", "ips.0.ip_address"),

					resource.TestCheckNoResourceAttr("data.scaleway_lb_ips.lbs_by_ip_address_other_zone", "ips.0.id"),
				),
			},
		},
	})
}
