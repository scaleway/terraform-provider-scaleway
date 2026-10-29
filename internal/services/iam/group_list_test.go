package iam_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccListIAMGroups_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccListIAMGroups_Basic because list resources are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCheckIamGroupDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_iam_group" "group1" {
					  name = "test-group-list-1"
					}
				`,
			},
			{
				Config: `
					resource "scaleway_iam_group" "group1" {
					  name = "test-group-list-1"
					}

					resource "scaleway_iam_group" "group2" {
					  name = "test-group-list-2"
					}
				`,
			},
			{
				Query: true,
				Config: `
					list "scaleway_iam_group" "by_name" {
					  provider = scaleway

					  config {
						organization_id = scaleway_iam_group.group1.organization_id
						name             = "test-group-list-1"
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_iam_group.by_name", 1),
				},
			},
			{
				Config: `
					resource "scaleway_iam_group" "group1" {
					  name = "test-group-list-1"
					}

					resource "scaleway_iam_group" "group2" {
					  name = "test-group-list-2"
					  tags = ["test-tag"]
					}
				`,
			},
			{
				Query: true,
				Config: `
					list "scaleway_iam_group" "by_tag" {
					  provider = scaleway

					  config {
						organization_id = scaleway_iam_group.group2.organization_id
						tag              = "test-tag"
					  }
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_iam_group.by_tag", 1),
				},
			},
		},
	})
}
