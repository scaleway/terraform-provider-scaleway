package vpc_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	accounttestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account/testfuncs"
)

func TestAccListVPCs_Basic(t *testing.T) {
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
					  project_id= scaleway_account_project.main.id
					  region = "fr-par"
					  name   = "test-vpc-fr-par"
					}
					
					resource "scaleway_vpc" "alt" {
					  project_id= scaleway_account_project.main.id
					  region = "nl-ams"
					  name   = "test-vpc-nl-ams"
					}

					resource "scaleway_vpc" "tag" {
					  project_id= scaleway_account_project.main.id
					  region = "fr-par"
                      tags = ["foobar"]
					}
`,
			},
			{
				Query: true,
				Config: `
					list "scaleway_vpc" "all" {
					  provider = scaleway
					
					  config {
						regions = ["*"]
						project_ids = [scaleway_account_project.main.id]
					  }
					}
					`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_vpc.all", 3),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_vpc" "fr-par" {
					  provider = scaleway
					
					  config {
						project_ids = [scaleway_account_project.main.id]
						regions = ["fr-par"]
					  }
					}
					`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_vpc.fr-par", 2),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_vpc" "by_name" {
					  provider = scaleway
					
					  config {
						project_ids = [scaleway_account_project.main.id]
						regions = ["*"]
						name = "test-vpc"
					  }
					}`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_vpc.by_name", 2),
				},
			},
			{
				Query: true,
				Config: `					
					list "scaleway_vpc" "by_tag" {
					  	provider = scaleway
					
						config {
							project_ids = [scaleway_account_project.main.id]
							regions = ["*"]
							tags = ["foobar"]
						}
					}`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_vpc.by_tag", 1),
				},
			},
		},
	})
}
