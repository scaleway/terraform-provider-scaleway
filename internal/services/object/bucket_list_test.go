package object_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
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

	bucketName1 := sdkacctest.RandomWithPrefix("tf-test-bucket-list-1")
	bucketName2 := sdkacctest.RandomWithPrefix("tf-test-bucket-list-2")
	bucketName3 := sdkacctest.RandomWithPrefix("tf-test-bucket-list-3")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             objectchecks.IsBucketDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
                resource "scaleway_object_bucket" "bucket1" {
                   name = "%s"
                }
             `, bucketName1),
			},

			{
				Config: fmt.Sprintf(`
                resource "scaleway_object_bucket" "bucket1" {
                   name = "%s"
                }

                resource "scaleway_object_bucket" "bucket2" {
                   name = "%s"
                   tags = {
                      environment = "test"
                   }
                }
             `, bucketName1, bucketName2),
			},

			{
				Config: fmt.Sprintf(`
                resource "scaleway_object_bucket" "bucket1" {
                   name = "%s"
                }

                resource "scaleway_object_bucket" "bucket2" {
                   name = "%s"
                   tags = {
                      environment = "test"
                   }
                }

                resource "scaleway_object_bucket" "bucket3" {
                   name = "%s"
                   region = "pl-waw"
                   tags = {
                      environment = "test"
                   }
                }
             `, bucketName1, bucketName2, bucketName3),
			},

			{
				Query: true,
				PreConfig: func() {
					ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
					defer cancel()

					err := retry.RetryContext(ctx, 2*time.Second, func() *retry.RetryError {
						return nil
					})
					if err != nil {
						t.Fatalf("error while checking for bucket:: %s", err)
					}
				},
				Config: fmt.Sprintf(`
					list "scaleway_object_bucket" "by_name" {
					   provider = scaleway

					   config {
						  regions     = ["*"]
						  project_ids = ["*"]
						  name        = "%s"
					   }
					}
				`, bucketName1),
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_object_bucket.by_name", 1),
				},
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
		},
	})
}
