package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayDataSourceVPCPublicGatewayIP_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayVPCPublicGatewayIPDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_vpc_public_gateway_ip" "main" {
					}`,
			},
			{
				Config: `
					resource "scaleway_vpc_public_gateway_ip" "main" {
					}

					data "scaleway_vpc_public_gateway_ip" "ip_by_id" {
						ip_id = "${scaleway_vpc_public_gateway_ip.main.id}"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayVPCPublicGatewayIPExists(tt, "scaleway_vpc_public_gateway_ip.main"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_vpc_public_gateway_ip.ip_by_id", "ip_id",
						"scaleway_vpc_public_gateway_ip.main", "id"),
				),
			},
		},
	})
}
