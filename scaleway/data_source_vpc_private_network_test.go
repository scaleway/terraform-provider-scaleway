package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayDataSourceVPCPrivateNetwork_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	pnName := "TestAccScalewayDataSourceVPCPrivateNetwork_Basic"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayVPCPrivateNetworkDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_vpc_private_network" "pn_test" {
					  name = "%s"
					}`, pnName),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_vpc_private_network" "pn_test" {
					  name = "%s"
					}

					data "scaleway_vpc_private_network" "pn_test_by_name" {
						name = "${scaleway_vpc_private_network.pn_test.name}"
					}

					data "scaleway_vpc_private_network" "pn_test_by_id" {
						private_network_id = "${scaleway_vpc_private_network.pn_test.id}"
					}
				`, pnName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayVPCPrivateNetworkExists(tt, "data.scaleway_vpc_private_network.pn_test"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_vpc_private_network.pn_test_by_name", "name",
						"scaleway_vpc_private_network.pn_test", "name"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_vpc_private_network.pn_test_by_id", "private_network_id",
						"scaleway_vpc_private_network.pn_test", "id"),
				),
			},
		},
	})
}
