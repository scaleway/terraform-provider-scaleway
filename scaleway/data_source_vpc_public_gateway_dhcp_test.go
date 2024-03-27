package scaleway_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccScalewayDataSourceVPCPublicGatewayDHCP_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayVPCPublicGatewayIPDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_vpc_public_gateway_dhcp" "main" {
						subnet = "192.168.1.0/24"
					}`,
			},
			{
				Config: `
					resource "scaleway_vpc_public_gateway_dhcp" "main" {
						subnet = "192.168.1.0/24"
					}

					data "scaleway_vpc_public_gateway_dhcp" "dhcp_by_id" {
						dhcp_id = "${scaleway_vpc_public_gateway_dhcp.main.id}"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayVPCPublicGatewayDHCPExists(tt, "scaleway_vpc_public_gateway_dhcp.main"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_vpc_public_gateway_dhcp.dhcp_by_id", "dhcp_id",
						"scaleway_vpc_public_gateway_dhcp.main", "id"),
				),
			},
		},
	})
}
