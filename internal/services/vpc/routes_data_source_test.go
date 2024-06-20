package vpc_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	vpcchecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/vpc/testfuncs"
)

func TestAccDataSourceRoutes_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      vpcchecks.CheckPrivateNetworkDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_vpc vpc01 {
						name = "tf-vpc-route-01"
                        enable_routing = true
					}

					resource scaleway_vpc_private_network pn01 {
						name = "tf-pn_route"
						ipv4_subnet {
							subnet = "172.16.64.0/22"
						}
						vpc_id = scaleway_vpc.vpc01.id
					}

					resource "scaleway_ipam_ip" "ip01" {
					  address = "172.16.64.7"
					  source {
						private_network_id = scaleway_vpc_private_network.pn01.id
					  }
					}`,
			},
			{
				Config: `
					resource scaleway_vpc vpc01 {
						name = "tf-vpc-route-01"
                        enable_routing = true
					}

					resource scaleway_vpc_private_network pn01 {
						name = "tf-pn_route"
						ipv4_subnet {
							subnet = "172.16.64.0/22"
						}
						vpc_id = scaleway_vpc.vpc01.id
					}

					resource "scaleway_ipam_ip" "ip01" {
					  address = "172.16.64.7"
					  source {
						private_network_id = scaleway_vpc_private_network.pn01.id
					  }
					}

					resource scaleway_vpc_public_gateway pg01 {
						name = "tf-gw-route"
						type = "VPC-GW-S"
					}

					resource scaleway_vpc_gateway_network main {
						gateway_id = scaleway_vpc_public_gateway.pg01.id
						private_network_id = scaleway_vpc_private_network.pn01.id
						enable_masquerade = true
						ipam_config {
							push_default_route = true
							ipam_ip_id = scaleway_ipam_ip.ip01.id
						}					
					}`,
			},
			{
				Config: `
					resource scaleway_vpc vpc01 {
						name = "tf-vpc-route-01"
					}

					resource scaleway_vpc_private_network pn01 {
						name = "tf-pn_route"
						ipv4_subnet {
							subnet = "172.16.64.0/22"
						}
						vpc_id = scaleway_vpc.vpc01.id
					}

					resource "scaleway_ipam_ip" "ip01" {
					  address = "172.16.64.7"
					  source {
						private_network_id = scaleway_vpc_private_network.pn01.id
					  }
					}

					resource scaleway_vpc_public_gateway pg01 {
						name = "tf-gw-route"
						type = "VPC-GW-S"
					}

					resource scaleway_vpc_gateway_network main {
						gateway_id = scaleway_vpc_public_gateway.pg01.id
						private_network_id = scaleway_vpc_private_network.pn01.id
						enable_masquerade = true
						ipam_config {
							push_default_route = true
							ipam_ip_id = scaleway_ipam_ip.ip01.id
						}					
					}

					data scaleway_vpc_routes routes_by_vpc_id {
						vpc_id = scaleway_vpc.vpc01.id
					}
					
					data scaleway_vpc_routes routes_by_ipv6 {
						vpc_id  = scaleway_vpc.vpc01.id
						is_ipv6 = true
					}

					data scaleway_vpc_routes routes_by_gw_network {
						vpc_id                = scaleway_vpc.vpc01.id
						nexthop_resource_type = "vpc_gateway_network"
					}
					`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaleway_vpc_routes.routes_by_vpc_id", "routes.#", "3"),
					resource.TestCheckResourceAttr("data.scaleway_vpc_routes.routes_by_ipv6", "routes.#", "1"),

					resource.TestCheckResourceAttr("data.scaleway_vpc_routes.routes_by_gw_network", "routes.#", "1"),
				),
			},
			{
				Config: `
					resource scaleway_vpc vpc01 {
						name = "tf-vpc-route-01"
					}

					resource scaleway_vpc_private_network pn01 {
						name = "tf-pn_route"
						ipv4_subnet {
							subnet = "172.16.64.0/22"
						}
						vpc_id = scaleway_vpc.vpc01.id
					}
					`,
			},
		},
	})
}
