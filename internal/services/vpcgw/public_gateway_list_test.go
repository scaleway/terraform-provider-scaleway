package vpcgw_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	vpcgwchecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/vpcgw/testfuncs"
)

func TestAccListPublicGateways_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccListPublicGateways_Basic because list resources are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             vpcgwchecks.IsGatewayDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_vpc_public_gateway" "gw1" {
					  zone = "fr-par-1"
					  name = "test-gw-list-1"
					  type = "VPC-GW-S"
					}

					resource "scaleway_vpc_public_gateway" "gw2" {
					  zone = "fr-par-1"
					  name = "test-gw-list-2"
					  type = "VPC-GW-S"
					  tags = ["test-gw-list-tagged"]
					}
`,
			},
			{
				Query: true,
				Config: `
					list "scaleway_vpc_public_gateway" "by_name" {
					  provider = scaleway

					  config {
						zones = ["fr-par-1"]
						name  = "test-gw-list-1"
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_vpc_public_gateway.by_name", 1),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_vpc_public_gateway" "by_tag" {
					  provider = scaleway

					  config {
						zones = ["fr-par-1"]
						tags  = ["test-gw-list-tagged"]
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_vpc_public_gateway.by_tag", 1),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_vpc_public_gateway" "by_type" {
					  provider = scaleway

					  config {
						zones = ["fr-par-1"]
						name  = "test-gw-list-2"
						types = ["VPC-GW-S"]
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_vpc_public_gateway.by_type", 1),
				},
			},
		},
	})
}
