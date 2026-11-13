package vpc_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccListVPCRoutes_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccListVPCRoutes_Basic because list resources are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_vpc" "test" {
					  region = "fr-par"
					  name   = "tf-route-list-vpc"
					}

					resource "scaleway_vpc_private_network" "target" {
					  vpc_id = scaleway_vpc.test.id
					  region = "fr-par"
					  name   = "tf-route-list-pn"
					}

					resource "scaleway_vpc_route" "test" {
					  vpc_id                     = scaleway_vpc.test.id
					  tags                       = ["tf-route-list-tag"]
					  nexthop_private_network_id = scaleway_vpc_private_network.target.id
					  destination                = "192.168.42.0/24"
					}
				`,
			},
			{
				Query: true,
				Config: `
					list "scaleway_vpc_route" "by_vpc" {
					  provider = scaleway

					  config {
					    regions = ["fr-par"]
					    vpc_id  = scaleway_vpc.test.id
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_vpc_route.by_vpc", 1),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_vpc_route" "by_tag" {
					  provider = scaleway

					  config {
					    regions = ["fr-par"]
					    vpc_id  = scaleway_vpc.test.id
					    tags    = ["tf-route-list-tag"]
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_vpc_route.by_tag", 1),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_vpc_route" "by_nexthop_pn" {
					  provider = scaleway

					  config {
					    regions                    = ["fr-par"]
					    vpc_id                     = scaleway_vpc.test.id
					    nexthop_private_network_id = scaleway_vpc_private_network.target.id
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_vpc_route.by_nexthop_pn", 1),
				},
			},
		},
	})
}
