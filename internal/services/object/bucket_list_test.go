package object_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	objectchecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/object/testfuncs"
)

func TestAccListObjectBuckets_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccListObjectBuckets_Basic because list resources are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	testDefaultRegion, _ := tt.Meta.ScwClient().GetDefaultRegion()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             objectchecks.IsBucketDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_object_bucket" "bucket1" {
						name = "test-bucket-list-1"
					}
				`,
			},

			{
				Config: `
					resource "scaleway_object_bucket" "bucket1" {
						name = "test-bucket-list-1"
					}

					resource "scaleway_object_bucket" "bucket2" {
						name = "test-bucket-list-2"
						tags = {
							environment = "test"
						}
					}
				`,
			},

			{
				Config: `
					resource "scaleway_object_bucket" "bucket1" {
						name = "test-bucket-list-1"
					}

					resource "scaleway_object_bucket" "bucket2" {
						name = "test-bucket-list-2"
						tags = {
							environment = "test"
						}
					}

					resource "scaleway_object_bucket" "bucket3" {
						name = "test-bucket-list-3"
						region = "pl-waw"
						tags = {
							environment = "test"
						}
					}
				`,
			},

			{
				Query: true,
				Config: `
					list "scaleway_object_bucket" "all" {
						provider = scaleway

						config {
							regions     = ["*"]
							project_ids = ["*"]
						}
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLengthAtLeast("list.scaleway_object_bucket.all", 3),
				},
			},

			{
				Query: true,
				Config: `
					list "scaleway_object_bucket" "by_region" {
						provider = scaleway

						config {
							regions     = [scaleway_object_bucket.bucket1.region]
							project_ids = [scaleway_object_bucket.bucket1.project_id]
						}
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLengthAtLeast("list.scaleway_object_bucket.by_region", 2),
				},
			},

			{
				Query: true,
				Config: `
					list "scaleway_object_bucket" "by_region_pl_waw" {
						provider = scaleway

						config {
							regions     = ["pl-waw"]
							project_ids = ["*"]
						}
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_object_bucket.by_region_pl_waw", 1),
				},
			},

			{
				Query: true,
				Config: fmt.Sprintf(`
					list "scaleway_object_bucket" "by_region_default" {
						provider = scaleway

						config {
							regions     = ["%s"]
							project_ids = ["*"]
						}
					}
				`, testDefaultRegion),
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLengthAtLeast("list.scaleway_object_bucket.by_region_default", 2),
				},
			},

			{
				Query: true,
				Config: `
					list "scaleway_object_bucket" "by_name" {
						provider = scaleway

						config {
							regions     = ["*"]
							project_ids = ["*"]
							name        = "test-bucket-list-1"
						}
					}
				`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_object_bucket.by_name", 1),
				},
			},
		},
	})
}
