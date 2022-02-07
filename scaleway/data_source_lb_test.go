package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayDataSourceLb_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayLbIPDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_lb_ip main {
					}

					resource scaleway_lb main {
					    ip_id = scaleway_lb_ip.main.id
						name = "data-test-lb"
						type = "LB-S"
					}
					
					data "scaleway_lb" "testByID" {
						lb_id = "${scaleway_lb.main.id}"
					}
					
					data "scaleway_lb" "testByName" {
						name = "${scaleway_lb.main.name}"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayLbExists(tt, "data.scaleway_lb.testByID"),
					testAccCheckScalewayLbExists(tt, "data.scaleway_lb.testByName"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_lb.testByID", "name",
						"scaleway_lb.main", "name"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_lb.testByName", "id",
						"scaleway_lb.main", "id"),
				),
			},
		},
	})
}
