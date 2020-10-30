package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayDataSourceLbIPBeta_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayLbIPBetaDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_lb_ip_beta" "test" {
					}
					
					data "scaleway_lb_ip_beta" "test" {
						ip_address = "${scaleway_lb_ip_beta.test.ip_address}"
					}
					
					data "scaleway_lb_ip_beta" "test2" {
						ip_id = "${scaleway_lb_ip_beta.test.id}"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayLbIPBetaExists(tt, "data.scaleway_lb_ip_beta.test"),
					resource.TestCheckResourceAttrPair("data.scaleway_lb_ip_beta.test", "ip_address", "scaleway_lb_ip_beta.test", "ip_address"),
					resource.TestCheckResourceAttrPair("data.scaleway_lb_ip_beta.test2", "ip_address", "scaleway_lb_ip_beta.test", "ip_address"),
				),
			},
		},
	})
}
