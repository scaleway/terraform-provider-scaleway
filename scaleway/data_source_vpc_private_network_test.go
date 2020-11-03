package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayDataSourceVPCPrivateNetwork_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	name := "TestAccScalewayDataSourceVPCPrivateNetwork_Basic"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayVPCPrivateNetworkDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_vpc_private_network" "main" {
						name = "%s"
					}
				`, name),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_vpc_private_network" "main" {
						name = "%s"
					}
					
					data "scaleway_vpc_private_network" "test" {
						name = "${scaleway_vpc_private_network.main.name}"
					}
					
					data "scaleway_vpc_private_network" "test2" {
						private_network_id = "${scaleway_vpc_private_network.main.id}"
					}
				`, name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayVPCPrivateNetworkExists(tt, "scaleway_vpc_private_network.main"),

					resource.TestCheckResourceAttr("data.scaleway_vpc_private_network.test", "name", name),
					resource.TestCheckResourceAttrSet("data.scaleway_vpc_private_network.test", "id"),

					resource.TestCheckResourceAttr("data.scaleway_vpc_private_network.test2", "name", name),
					resource.TestCheckResourceAttrSet("data.scaleway_vpc_private_network.test2", "id"),
				),
			},
		},
	})
}
