package vpc_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	vpctestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/vpc/testfuncs"
)

func TestAccVPCIngressRule_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             vpctestfuncs.CheckIngressRuleDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_vpc" "vpc01" {
					  name = "tf-ingress-vpc"
					}

					resource "scaleway_vpc_private_network" "pn01" {
					  name   = "tf-ingress-pn"
					  vpc_id = scaleway_vpc.vpc01.id
					}

					resource "scaleway_ipam_ip" "ip01" {
				 	  source {
						private_network_id = scaleway_vpc_private_network.pn01.id
					  }
					}

					resource "scaleway_vpc_ingress_rule" "main" {
					  vpc_id                     = scaleway_vpc.vpc01.id
					  source                     = "10.0.0.0/24"
					  nexthop_private_network_id = scaleway_vpc_private_network.pn01.id
					  nexthop_resource_ip        = scaleway_ipam_ip.ip01.address
					  description                = "ingress rule basic"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					vpctestfuncs.IsIngressRulePresent(tt, "scaleway_vpc_ingress_rule.main"),
					resource.TestCheckResourceAttr("scaleway_vpc_ingress_rule.main", "source", "10.0.0.0/24"),
					resource.TestCheckResourceAttr("scaleway_vpc_ingress_rule.main", "description", "ingress rule basic"),
					resource.TestCheckResourceAttrPair("scaleway_vpc_ingress_rule.main", "vpc_id", "scaleway_vpc.vpc01", "id"),
					resource.TestCheckResourceAttrPair("scaleway_vpc_ingress_rule.main", "nexthop_private_network_id", "scaleway_vpc_private_network.pn01", "id"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_ingress_rule.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_ingress_rule.main", "updated_at"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_ingress_rule.main", "is_ipv6"),
					resource.TestCheckResourceAttr("scaleway_vpc_ingress_rule.main", "region", "fr-par"),
				),
			},
			{
				Config: `
					resource "scaleway_vpc" "vpc01" {
					  name = "tf-ingress-vpc"
					}

					resource "scaleway_vpc_private_network" "pn01" {
					  name   = "tf-ingress-pn"
					  vpc_id = scaleway_vpc.vpc01.id
					}

					resource "scaleway_ipam_ip" "ip01" {
				 	  source {
						private_network_id = scaleway_vpc_private_network.pn01.id
					  }
					}

					resource "scaleway_vpc_ingress_rule" "main" {
					  vpc_id                     = scaleway_vpc.vpc01.id
					  source                     = "10.0.0.0/24"
					  nexthop_private_network_id = scaleway_vpc_private_network.pn01.id
					  nexthop_resource_ip        = scaleway_ipam_ip.ip01.address
					  description                = "ingress rule updated"
					  tags                       = ["terraform", "ingress"]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					vpctestfuncs.IsIngressRulePresent(tt, "scaleway_vpc_ingress_rule.main"),
					resource.TestCheckResourceAttr("scaleway_vpc_ingress_rule.main", "description", "ingress rule updated"),
					resource.TestCheckResourceAttr("scaleway_vpc_ingress_rule.main", "tags.#", "2"),
					resource.TestCheckResourceAttr("scaleway_vpc_ingress_rule.main", "tags.0", "terraform"),
					resource.TestCheckResourceAttr("scaleway_vpc_ingress_rule.main", "tags.1", "ingress"),
				),
			},
		},
	})
}
