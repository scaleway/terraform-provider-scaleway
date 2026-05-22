package vpc_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	accounttestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account/testfuncs"
)

func TestAccListVPCConnectors_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccListVPCConnectors_Basic because list resources are not yet supported on OpenTofu")
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

					resource "scaleway_vpc" "source" {
					  project_id = scaleway_account_project.main.id
					  region     = "fr-par"
					  name       = "tf-vpc-source"
					}

					resource "scaleway_vpc" "target" {
					  project_id = scaleway_account_project.main.id
					  region     = "fr-par"
					  name       = "tf-vpc-target"
					}

					resource "scaleway_vpc_connector" "main" {
					  name          = "tf-conn-list"
					  vpc_id        = scaleway_vpc.source.id
					  target_vpc_id = scaleway_vpc.target.id
					  tags          = ["tf-conn-list-tag"]
					}
				`,
			},
			{
				Query: true,
				Config: `
					list "scaleway_vpc_connector" "all" {
					  provider = scaleway

					  config {
					    regions     = ["*"]
					    project_ids = [scaleway_account_project.main.id]
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_vpc_connector.all", 1),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_vpc_connector" "by_name" {
					  provider = scaleway

					  config {
					    regions     = ["fr-par"]
					    project_ids = [scaleway_account_project.main.id]
					    name        = "tf-conn-list"
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_vpc_connector.by_name", 1),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_vpc_connector" "by_vpc" {
					  provider = scaleway

					  config {
					    regions     = ["fr-par"]
					    project_ids = [scaleway_account_project.main.id]
					    vpc_id      = scaleway_vpc.source.id
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_vpc_connector.by_vpc", 1),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_vpc_connector" "by_tag" {
					  provider = scaleway

					  config {
					    regions     = ["fr-par"]
					    project_ids = [scaleway_account_project.main.id]
					    tags        = ["tf-conn-list-tag"]
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_vpc_connector.by_tag", 1),
				},
			},
		},
	})
}
