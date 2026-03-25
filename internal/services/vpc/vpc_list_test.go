package vpc_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccListVPCs_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Create VPCs in different regions with different tags
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
					}`,
			},
			{
				Query: true,
				Config: `
					list "scaleway_vpc" "all" {
					  provider = scaleway
					
					  config {
						region = "all"
						project_id = scaleway_account_project.main.id
					  }
					}
					`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					// Check that we can list all VPCs
					querycheck.ExpectLength("list.scaleway_vpc.all", 2),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_vpc" "fr-par" {
					  provider = scaleway
					
					  config {
						project_id = scaleway_account_project.main.id
						region = "fr-par"
					  }
					}
					`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					intercept{},

					// Check that we can filter by region
					querycheck.ExpectLength("list.scaleway_vpc.fr-par", 1),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_vpc" "by_name" {
					  provider = scaleway
					
					  config {
						project_id = scaleway_account_project.main.id
						region = "all"
						name = "test-vpc"
					  }
					}`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					// Check that we can filter by name pattern
					querycheck.ExpectLength("list.scaleway_vpc.by_name", 2),
				},
			},
		},
	})
}

type intercept struct{}

func (i intercept) CheckQuery(ctx context.Context, request querycheck.CheckQueryRequest, response *querycheck.CheckQueryResponse) {
	return
}
