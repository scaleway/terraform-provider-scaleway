package vpc_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	accounttestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account/testfuncs"
)

func TestAccListPrivateNetworks_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccListPrivateNetworks_Basic because list resources are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             accounttestfuncs.IsProjectDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_account_project" "main" {}

					resource "scaleway_vpc" "main" {
					  project_id = scaleway_account_project.main.id
					  region     = "fr-par"
					  name       = "test-vpc-pn-list"
					}

					resource "scaleway_vpc_private_network" "pn1" {
					  project_id = scaleway_account_project.main.id
					  vpc_id     = scaleway_vpc.main.id
					  region     = "fr-par"
					  name       = "test-pn-1"
					}

					resource "scaleway_vpc_private_network" "pn2" {
					  project_id = scaleway_account_project.main.id
					  vpc_id     = scaleway_vpc.main.id
					  region     = "fr-par"
					  name       = "test-pn-2"
					  tags       = ["tagged"]
					}
`,
			},
			{
				Query: true,
				Config: `
					list "scaleway_vpc_private_network" "all" {
					  provider = scaleway

					  config {
						regions     = ["fr-par"]
						project_ids = [scaleway_account_project.main.id]
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_vpc_private_network.all", 2),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_vpc_private_network" "by_name" {
					  provider = scaleway

					  config {
						regions     = ["fr-par"]
						project_ids = [scaleway_account_project.main.id]
						name        = "test-pn-1"
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_vpc_private_network.by_name", 1),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_vpc_private_network" "by_tag" {
					  provider = scaleway

					  config {
						regions     = ["fr-par"]
						project_ids = [scaleway_account_project.main.id]
						tags        = ["tagged"]
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_vpc_private_network.by_tag", 1),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_vpc_private_network" "by_vpc" {
					  provider = scaleway

					  config {
						regions     = ["fr-par"]
						project_ids = [scaleway_account_project.main.id]
						vpc_id      = scaleway_vpc.main.id
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_vpc_private_network.by_vpc", 2),
				},
			},
		},
	})
}
