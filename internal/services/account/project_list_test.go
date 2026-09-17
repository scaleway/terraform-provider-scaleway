package account_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	accounttestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account/testfuncs"
)

func TestAccListResource_Project(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccListResource_Project because list resources are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	orgID, orgIDExists := tt.Meta.ScwClient().GetDefaultOrganizationID()
	if !orgIDExists {
		orgID = "105bdce1-64c0-48ab-899d-868455867ecf"
	}

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Query: true,
				// lintignore:AT004
				Config: fmt.Sprintf(`
					provider "scaleway" {}

					list "scaleway_account_project" "all" {
						provider = scaleway

						config {
							organization_id = "%s"
						}
					}
				`, orgID),
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLengthAtLeast("list.scaleway_account_project.all", 1),
				},
			},
		},
	})
}

func TestAccListResource_Project_ByName(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccListResource_Project_ByName because list resources are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	orgID, orgIDExists := tt.Meta.ScwClient().GetDefaultOrganizationID()
	if !orgIDExists {
		orgID = "105bdce1-64c0-48ab-899d-868455867ecf"
	}

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             accounttestfuncs.IsProjectDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_account_project" "project1" {
						name = "tf-tests-project-list-by-name"
						description = "Project for testing list by name"
					}
				`,
			},
			{
				Config: `
					resource "scaleway_account_project" "project1" {
						name = "tf-tests-project-list-by-name"
						description = "Project for testing list by name"
					}

					resource "scaleway_account_project" "project2" {
						name = "tf-tests-project-list-other"
						description = "Another project"
					}
				`,
			},
			{
				Query: true,
				Config: fmt.Sprintf(`
					list "scaleway_account_project" "by_name" {
						provider = scaleway

						config {
							organization_id = "%s"
							name = "tf-tests-project-list-by-name"
						}
					}
				`, orgID),
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_account_project.by_name", 1),
				},
			},
		},
	})
}

func TestAccListResource_Project_ByProjectIDs(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccListResource_Project_ByProjectIDs because list resources are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	orgID, orgIDExists := tt.Meta.ScwClient().GetDefaultOrganizationID()
	if !orgIDExists {
		orgID = "105bdce1-64c0-48ab-899d-868455867ecf"
	}

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             accounttestfuncs.IsProjectDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_account_project" "project1" {
						name = "tf-tests-project-list-id-1"
						description = "First project for testing list by IDs"
					}
				`,
			},
			{
				Config: `
					resource "scaleway_account_project" "project1" {
						name = "tf-tests-project-list-id-1"
						description = "First project for testing list by IDs"
					}

					resource "scaleway_account_project" "project2" {
						name = "tf-tests-project-list-id-2"
						description = "Second project for testing list by IDs"
					}
				`,
			},
			{
				Query: true,
				Config: fmt.Sprintf(`
					list "scaleway_account_project" "by_project_ids" {
						provider = scaleway

						config {
							organization_id = "%s"
							project_ids = [
								scaleway_account_project.project1.id,
								scaleway_account_project.project2.id,
							]
						}
					}
				`, orgID),
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_account_project.by_project_ids", 2),
				},
			},
		},
	})
}
