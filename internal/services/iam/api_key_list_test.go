package iam_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccListIAMAPIKeys_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccListIAMAPIKeys_Basic because list resources are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckIamAPIKeyDestroy(tt),
			testAccCheckIamApplicationDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_iam_application" "main" {
						name = "tf_tests_app_key_basic"
					}

					resource "scaleway_iam_api_key" "key1" {
						application_id = scaleway_iam_application.main.id
						description = "test-api-key-list-1"
					}
				`,
			},
			{
				Config: `
					resource "scaleway_iam_application" "main" {
						name = "tf_tests_app_key_basic"
					}

					resource "scaleway_iam_api_key" "key1" {
						application_id = scaleway_iam_application.main.id
						description = "test-api-key-list-1"
					}
				`,
			},
			{
				Query: true,
				Config: `
					list "scaleway_iam_api_key" "all" {
						provider = scaleway

						config {
							organization_id = scaleway_iam_application.main.organization_id
						}
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLengthAtLeast("list.scaleway_iam_api_key.all", 1),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_iam_api_key" "by_description" {
						provider = scaleway

						config {
							organization_id = scaleway_iam_application.main.organization_id
							description      = "test-api-key-list-1"
						}
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_iam_api_key.by_description", 1),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_iam_api_key" "by_editable" {
						provider = scaleway

						config {
							organization_id = scaleway_iam_application.main.organization_id
							editable        = true
						}
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLengthAtLeast("list.scaleway_iam_api_key.by_editable", 1),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_iam_api_key" "by_expired" {
						provider = scaleway

						config {
							organization_id = scaleway_iam_application.main.organization_id
							expired          = false
						}
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLengthAtLeast("list.scaleway_iam_api_key.by_expired", 1),
				},
			},
			{
				Query: true,
				Config: `
					list "scaleway_iam_api_key" "by_bearer_id" {
						provider = scaleway

						config {
							organization_id = scaleway_iam_application.main.organization_id
							bearer_id       = scaleway_iam_application.main.id
							bearer_type 	= "application"
						}
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_iam_api_key.by_bearer_id", 1),
				},
			},
		},
	})
}
