package lb_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	vpcchecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/vpc/testfuncs"
)

func TestAccLBPrivateNetwork_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			isLbDestroyed(tt),
			vpcchecks.CheckPrivateNetworkDestroy(tt),
			vpcchecks.CheckVPCDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
				resource "scaleway_vpc" "vpc01" {
				  name = "my vpc"
				}
				
				resource "scaleway_vpc_private_network" "pn01" {
				  vpc_id = scaleway_vpc.vpc01.id
				  ipv4_subnet {
					subnet = "172.16.32.0/22"
				  }
				}
				
				resource "scaleway_ipam_ip" "ip01" {
				  address = "172.16.32.7"
				  source {
					private_network_id = scaleway_vpc_private_network.pn01.id
				  }
				}
				
				resource "scaleway_lb" "lb01" {
				  name = "test-lb-private-network"
				  type = "LB-S"
				}

				resource "scaleway_lb_private_network" "lbpn01" {
				  lb_id                 = scaleway_lb.lb01.id
				  private_network_id    = scaleway_vpc_private_network.pn01.id
				  ipam_ip_ids           = [scaleway_ipam_ip.ip01.id]
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"scaleway_lb.lb01", "id",
						"scaleway_lb_private_network.lbpn01", "lb_id"),
					resource.TestCheckResourceAttrPair(
						"scaleway_vpc_private_network.pn01", "id",
						"scaleway_lb_private_network.lbpn01", "private_network_id"),
					resource.TestCheckResourceAttrPair(
						"scaleway_ipam_ip.ip01", "id",
						"scaleway_lb_private_network.lbpn01", "ipam_ip_ids.0"),
				),
			},
		},
	})
}
