package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayDataSourceLbBackend_Basic(t *testing.T) {
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
						name = "data-test-lb-backend"
						type = "LB-S"
					}

					resource "scaleway_lb_backend" "main" {
					  lb_id            = scaleway_lb.main.id
					  name             = "backend01"
					  forward_protocol = "http"
					  forward_port     = "80"
					}
					
					data "scaleway_lb_backend" "byID" {
						backend_id = "${scaleway_lb_backend.main.id}"
					}
					
					data "scaleway_lb_backend" "byName" {
						name = "${scaleway_lb_backend.main.name}"
						lb_id = "${scaleway_lb.main.id}"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.scaleway_lb_backend.byID", "name",
						"scaleway_lb_backend.main", "name"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_lb_backend.byName", "id",
						"scaleway_lb_backend.main", "id"),
				),
			},
		},
	})
}
