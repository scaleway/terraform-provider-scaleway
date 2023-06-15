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
					testAccCheckScalewayVPCPrivateNetworkExists(tt, "scaleway_vpc_private_network.pn_test"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_vpc_private_network.pn_test_by_name", "name",
						"scaleway_vpc_private_network.pn_test", "name"),
					resource.TestCheckResourceAttr(
						"data.scaleway_vpc_private_network.pn_test_by_name",
						"ipv4_subnet.#", "1"),
					resource.TestCheckResourceAttr(
						"data.scaleway_vpc_private_network.pn_test_by_name",
						"ipv6_subnets.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(
						"data.scaleway_vpc_private_network.pn_test_by_name", "ipv4_subnet.0.subnet",
						"scaleway_vpc_private_network.pn_test", "ipv4_subnet.0.subnet"),
					resource.TestCheckTypeSetElemAttrPair(
						"data.scaleway_vpc_private_network.pn_test_by_name", "ipv6_subnets.0.subnet",
						"scaleway_vpc_private_network.pn_test", "ipv6_subnets.0.subnet"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_vpc_private_network.pn_test_by_id", "private_network_id",
						"scaleway_vpc_private_network.pn_test", "id"),
					resource.TestCheckResourceAttr(
						"data.scaleway_vpc_private_network.pn_test_by_id",
						"ipv4_subnet.#", "1"),
					resource.TestCheckResourceAttr(
						"data.scaleway_vpc_private_network.pn_test_by_id",
						"ipv6_subnets.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(
						"data.scaleway_vpc_private_network.pn_test_by_id", "ipv4_subnet.0.subnet",
						"scaleway_vpc_private_network.pn_test", "ipv4_subnet.0.subnet"),
					resource.TestCheckTypeSetElemAttrPair(
						"data.scaleway_vpc_private_network.pn_test_by_id", "ipv6_subnets.0.subnet",
						"scaleway_vpc_private_network.pn_test", "ipv6_subnets.0.subnet"),
				),
			},
		},
	})
}

func TestAccScalewayDataSourceVPCPrivateNetwork_Regional(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	pnName := "TestAccScalewayDataSourceVPCPrivateNetwork_Regional"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayVPCPrivateNetworkDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_vpc_private_network" "pn_test" {
					  name = "%s"
					  is_regional = true
					}`, pnName),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_vpc_private_network" "pn_test" {
					  name = "%s"
					  is_regional = true
					}

					data "scaleway_vpc_private_network" "pn_test_by_name" {
						name = scaleway_vpc_private_network.pn_test.name
						is_regional = scaleway_vpc_private_network.pn_test.is_regional
					}

					data "scaleway_vpc_private_network" "pn_test_by_id" {
						private_network_id = scaleway_vpc_private_network.pn_test.id
						is_regional = scaleway_vpc_private_network.pn_test.is_regional
					}
				`, pnName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayVPCPrivateNetworkExists(tt, "scaleway_vpc_private_network.pn_test"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_vpc_private_network.pn_test_by_name", "name",
						"scaleway_vpc_private_network.pn_test", "name"),
					resource.TestCheckResourceAttr(
						"data.scaleway_vpc_private_network.pn_test_by_name",
						"ipv4_subnet.#", "1"),
					resource.TestCheckResourceAttr(
						"data.scaleway_vpc_private_network.pn_test_by_name",
						"ipv6_subnets.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(
						"data.scaleway_vpc_private_network.pn_test_by_name", "ipv4_subnet.0.subnet",
						"scaleway_vpc_private_network.pn_test", "ipv4_subnet.0.subnet"),
					resource.TestCheckTypeSetElemAttrPair(
						"data.scaleway_vpc_private_network.pn_test_by_name", "ipv6_subnets.0.subnet",
						"scaleway_vpc_private_network.pn_test", "ipv6_subnets.0.subnet"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_vpc_private_network.pn_test_by_id", "private_network_id",
						"scaleway_vpc_private_network.pn_test", "id"),
					resource.TestCheckResourceAttr(
						"data.scaleway_vpc_private_network.pn_test_by_id",
						"ipv4_subnet.#", "1"),
					resource.TestCheckResourceAttr(
						"data.scaleway_vpc_private_network.pn_test_by_id",
						"ipv6_subnets.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(
						"data.scaleway_vpc_private_network.pn_test_by_id", "ipv4_subnet.0.subnet",
						"scaleway_vpc_private_network.pn_test", "ipv4_subnet.0.subnet"),
					resource.TestCheckTypeSetElemAttrPair(
						"data.scaleway_vpc_private_network.pn_test_by_id", "ipv6_subnets.0.subnet",
						"scaleway_vpc_private_network.pn_test", "ipv6_subnets.0.subnet"),
				),
			},
		},
	})
}
