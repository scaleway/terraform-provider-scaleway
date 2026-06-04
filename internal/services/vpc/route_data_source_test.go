package vpc_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceRoute_ByID(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isRouteDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_vpc" "vpc01" {
					  name = "tf-vpc-ds-route"
					}

					resource "scaleway_vpc_private_network" "pn01" {
					  name = "tf-pn-ds-route"
					  ipv4_subnet {
					    subnet = "172.16.32.0/22"
					  }
					  vpc_id = scaleway_vpc.vpc01.id
					}

					resource "scaleway_instance_server" "server01" {
					  name  = "tf-server-ds-route"
					  type  = "PLAY2-MICRO"
					  image = "openvpn"
					}

					resource "scaleway_instance_private_nic" "pnic01" {
					  private_network_id = scaleway_vpc_private_network.pn01.id
					  server_id          = scaleway_instance_server.server01.id
					}

					resource "scaleway_vpc_route" "rt01" {
					  vpc_id                     = scaleway_vpc.vpc01.id
					  description                = "tf-route-ds"
					  tags                       = ["tf", "route", "ds"]
					  destination                = "10.0.0.0/24"
					  nexthop_resource_id        = scaleway_instance_private_nic.pnic01.id
					  nexthop_private_network_id = scaleway_vpc_private_network.pn01.id
					}

					data "scaleway_vpc_route" "by_id" {
					  route_id = scaleway_vpc_route.rt01.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isRoutePresent(tt, "scaleway_vpc_route.rt01"),
					resource.TestCheckResourceAttrPair("data.scaleway_vpc_route.by_id", "route_id", "scaleway_vpc_route.rt01", "id"),
					resource.TestCheckResourceAttr("data.scaleway_vpc_route.by_id", "destination", "10.0.0.0/24"),
					resource.TestCheckResourceAttr("data.scaleway_vpc_route.by_id", "description", "tf-route-ds"),
					resource.TestCheckResourceAttr("data.scaleway_vpc_route.by_id", "tags.#", "3"),
					resource.TestCheckResourceAttr("data.scaleway_vpc_route.by_id", "tags.0", "tf"),
					resource.TestCheckResourceAttr("data.scaleway_vpc_route.by_id", "tags.1", "route"),
					resource.TestCheckResourceAttr("data.scaleway_vpc_route.by_id", "tags.2", "ds"),
				),
			},
		},
	})
}

func TestAccDataSourceRoute_ByFilters(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isRouteDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_vpc" "vpc01" {
					  name = "tf-vpc-ds-route-filter"
					}

					resource "scaleway_vpc_private_network" "pn01" {
					  name = "tf-pn-ds-route-filter"
					  ipv4_subnet {
					    subnet = "172.16.32.0/22"
					  }
					  vpc_id = scaleway_vpc.vpc01.id
					}

					resource "scaleway_instance_server" "server01" {
					  name  = "tf-server-ds-route-filter"
					  type  = "PLAY2-MICRO"
					  image = "openvpn"
					}

					resource "scaleway_instance_private_nic" "pnic01" {
					  private_network_id = scaleway_vpc_private_network.pn01.id
					  server_id          = scaleway_instance_server.server01.id
					}

					resource "scaleway_vpc_route" "rt01" {
					  vpc_id                     = scaleway_vpc.vpc01.id
					  description                = "tf-route-ds-filter"
					  tags                       = ["tf", "ds-filter"]
					  destination                = "10.0.0.0/24"
					  nexthop_resource_id        = scaleway_instance_private_nic.pnic01.id
					  nexthop_private_network_id = scaleway_vpc_private_network.pn01.id
					}

					data "scaleway_vpc_route" "by_tags" {
					  vpc_id     = scaleway_vpc.vpc01.id
					  tags       = ["tf", "ds-filter"]
					  depends_on = [scaleway_vpc_route.rt01]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isRoutePresent(tt, "scaleway_vpc_route.rt01"),
					resource.TestCheckResourceAttrPair("data.scaleway_vpc_route.by_tags", "route_id", "scaleway_vpc_route.rt01", "id"),
					resource.TestCheckResourceAttr("data.scaleway_vpc_route.by_tags", "destination", "10.0.0.0/24"),
					resource.TestCheckResourceAttr("data.scaleway_vpc_route.by_tags", "description", "tf-route-ds-filter"),
				),
			},
		},
	})
}
