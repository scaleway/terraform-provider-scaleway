package iam_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccListIAMApplications_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccListIAMApplications_Basic because list resources are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCheckIamApplicationDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_iam_application" "app1" {
						name = "test-app-list-1"
					}
				`,
			},
			{
				Config: `
					resource "scaleway_iam_application" "app1" {
						name = "test-app-list-1"
					}

					resource "scaleway_iam_application" "app2" {
						name = "test-app-list-2"
					}
				`,
			},
			{
				Query: true,
				Config: `
					list "scaleway_iam_application" "all" {
						provider = scaleway

						config {
							organization_id = scaleway_iam_application.app1.organization_id
						}
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLengthAtLeast("list.scaleway_iam_application.all", 2),
				},
			},
			{
				Config: `
					resource "scaleway_iam_application" "app1" {
						name = "test-app-list-1"
						tags = ["toto"]
					}

					resource "scaleway_iam_application" "app2" {
						name = "test-app-list-2"
						tags = ["test-tag"]
					}
				`,
			},
			{
				Query: true,
				Config: `
					list "scaleway_iam_application" "by_tag" {
						provider = scaleway

						config {
							organization_id = scaleway_iam_application.app2.organization_id
							tag              = "test-tag"
						}
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_iam_application.by_tag", 1),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_iam_application" "by_name" {
						provider = scaleway

						config {
							organization_id = scaleway_iam_application.app1.organization_id
							name              = "test-app-list-1"
						}
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_iam_application.by_name", 1),
				},
			},
		},
	})
}
