package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayDataSourceVPCPrivateNetwork_Basic(t *testing.T) {
	pnName := acctest.RandString(10)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayVPCPrivateNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_vpc_private_network" "pn_test" {
					  name = "` + pnName + `"
					}

					data "scaleway_vpc_private_network" "pn_test_by_name" {
						name = "${scaleway_vpc_private_network.pn_test.name}"
					}

					data "scaleway_vpc_private_network" "pn_test_by_id" {
						private_network_id = "${scaleway_vpc_private_network.pn_test.id}"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayVPCPrivateNetworkExists("data.scaleway_vpc_private_network.pn_test"),
					resource.TestCheckResourceAttrPair("data.scaleway_vpc_private_network.pn_test_by_name", "name", "scaleway_vpc_private_network.pn_test", "name"),
					resource.TestCheckResourceAttrPair("data.scaleway_vpc_private_network.pn_test_by_id", "private_network_id", "scaleway_vpc_private_network.pn_test", "id"),
				),
			},
		},
	})
}
