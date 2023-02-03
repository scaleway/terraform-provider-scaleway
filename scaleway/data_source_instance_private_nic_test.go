package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayDataSourceInstancePrivateNIC_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayInstancePrivateNICDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_server" "server" {
						name = "test-terraform-datasource-private-nic"
						type = "DEV1-S"
						image = "ubuntu_jammy"
					}
					resource "scaleway_vpc_private_network" "vpc" {}

					resource "scaleway_instance_private_nic" "nic" {
						server_id = scaleway_instance_server.server.id
						private_network_id = scaleway_vpc_private_network.vpc.id
					}`,
			},
			{
				Config: `
					resource "scaleway_instance_server" "server" {
						name = "test-terraform-datasource-private-nic"
						type = "DEV1-S"
						image = "ubuntu_jammy"
					}
					resource "scaleway_vpc_private_network" "vpc" {}

					resource "scaleway_instance_private_nic" "nic" {
						server_id = scaleway_instance_server.server.id
						private_network_id = scaleway_vpc_private_network.vpc.id
					}

					data scaleway_instance_private_nic find_by_nic_id {
						server_id = scaleway_instance_server.server.id
						private_nic_id = split("/", scaleway_instance_private_nic.nic.id)[2]
					}

					data scaleway_instance_private_nic find_by_vpc_id {
						server_id = scaleway_instance_server.server.id
						private_network_id = scaleway_vpc_private_network.vpc.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstancePrivateNICExists(tt, "scaleway_instance_private_nic.nic"),

					resource.TestCheckResourceAttrPair("data.scaleway_instance_private_nic.find_by_nic_id", "id", "data.scaleway_instance_private_nic.find_by_vpc_id", "id"),
					resource.TestCheckResourceAttrSet("data.scaleway_instance_private_nic.find_by_nic_id", "mac_address"),
					resource.TestCheckResourceAttrSet("data.scaleway_instance_private_nic.find_by_vpc_id", "mac_address"),
				),
			},
		},
	})
}
