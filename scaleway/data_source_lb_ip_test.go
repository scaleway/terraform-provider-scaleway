package scaleway_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccScalewayDataSourceLbIP_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.TestAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayLbIPDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_lb_ip" "test" {
					}
					
					data "scaleway_lb_ip" "test" {
						ip_address = "${scaleway_lb_ip.test.ip_address}"
					}
					
					data "scaleway_lb_ip" "test2" {
						ip_id = "${scaleway_lb_ip.test.id}"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayLbIPExists(tt, "data.scaleway_lb_ip.test"),
					resource.TestCheckResourceAttrPair("data.scaleway_lb_ip.test", "ip_address", "scaleway_lb_ip.test", "ip_address"),
					resource.TestCheckResourceAttrPair("data.scaleway_lb_ip.test2", "ip_address", "scaleway_lb_ip.test", "ip_address"),
				),
			},
		},
	})
}
