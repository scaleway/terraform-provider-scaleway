package vpc_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	vpctestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/vpc/testfuncs"
)

func TestAccDataSourceVPCIngressRule_ByID(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             vpctestfuncs.CheckIngressRuleDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_vpc" "vpc01" {
					  name = "tf-ingress-dsid-vpc"
					}

					resource "scaleway_vpc_private_network" "pn01" {
					  name   = "tf-ingress-dsid-pn"
					  vpc_id = scaleway_vpc.vpc01.id
					}

					resource "scaleway_ipam_ip" "ip01" {
				 	  source {
						private_network_id = scaleway_vpc_private_network.pn01.id
					  }
					}

					resource "scaleway_vpc_ingress_rule" "main" {
					  vpc_id                     = scaleway_vpc.vpc01.id
					  source                     = "10.1.0.0/24"
					  nexthop_private_network_id = scaleway_vpc_private_network.pn01.id
					  nexthop_resource_ip        = scaleway_ipam_ip.ip01.address
					  description                = "ingress rule data source by id"
					}

					data "scaleway_vpc_ingress_rule" "by_id" {
					  ingress_rule_id = scaleway_vpc_ingress_rule.main.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					vpctestfuncs.IsIngressRulePresent(tt, "scaleway_vpc_ingress_rule.main"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_vpc_ingress_rule.by_id", "source",
						"scaleway_vpc_ingress_rule.main", "source"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_vpc_ingress_rule.by_id", "vpc_id",
						"scaleway_vpc_ingress_rule.main", "vpc_id"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_vpc_ingress_rule.by_id", "nexthop_private_network_id",
						"scaleway_vpc_ingress_rule.main", "nexthop_private_network_id"),
				),
			},
		},
	})
}

func TestAccDataSourceVPCIngressRule_ByFilters(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             vpctestfuncs.CheckIngressRuleDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_vpc" "vpc01" {
					  name = "tf-ingress-dsfilt-vpc"
					}

					resource "scaleway_vpc_private_network" "pn01" {
					  name   = "tf-ingress-dsfilt-pn"
					  vpc_id = scaleway_vpc.vpc01.id
					}

					resource "scaleway_ipam_ip" "ip01" {
				 	  source {
						private_network_id = scaleway_vpc_private_network.pn01.id
					  }
					}

					resource "scaleway_vpc_ingress_rule" "main" {
					  vpc_id                     = scaleway_vpc.vpc01.id
					  source                     = "10.2.0.0/24"
					  nexthop_private_network_id = scaleway_vpc_private_network.pn01.id
					  nexthop_resource_ip        = scaleway_ipam_ip.ip01.address
					  description                = "ingress rule data source by filters"
					}

					data "scaleway_vpc_ingress_rule" "by_pn" {
					  nexthop_private_network_id = scaleway_vpc_ingress_rule.main.nexthop_private_network_id
					  depends_on                 = [scaleway_vpc_ingress_rule.main]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.scaleway_vpc_ingress_rule.by_pn", "id",
						"scaleway_vpc_ingress_rule.main", "id"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_vpc_ingress_rule.by_pn", "vpc_id",
						"scaleway_vpc_ingress_rule.main", "vpc_id"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_vpc_ingress_rule.by_pn", "source",
						"scaleway_vpc_ingress_rule.main", "source"),
				),
			},
		},
	})
}
