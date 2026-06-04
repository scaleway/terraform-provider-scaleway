package object_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func testAccCheckBucketServerSideEncryptionConfigurationDestroy(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		// The bucket server side encryption configuration is tied to the bucket
		// If the bucket is destroyed, the configuration is automatically destroyed
		return nil
	}
}

func TestAccDataSourceBucketServerSideEncryptionConfiguration_ByID(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCheckBucketServerSideEncryptionConfigurationDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_object_bucket" "main" {
					  name = "test-acc-scaleway-object-bucket-ds-sse-id"
					  region = "fr-par"
					}

					resource "scaleway_object_bucket_server_side_encryption_configuration" "main" {
					  bucket = scaleway_object_bucket.main.name
					  rule {
						apply_server_side_encryption_by_default {
						  sse_algorithm = "AES256"
						}
					  }
					}

					data "scaleway_object_bucket_server_side_encryption_configuration" "by_id" {
					  bucket_server_side_encryption_configuration_id = scaleway_object_bucket_server_side_encryption_configuration.main.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketServerSideEncryptionConfigurationExists(tt, "scaleway_object_bucket_server_side_encryption_configuration.main"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_object_bucket_server_side_encryption_configuration.by_id", "bucket",
						"scaleway_object_bucket_server_side_encryption_configuration.main", "bucket"),
					resource.TestCheckResourceAttr("data.scaleway_object_bucket_server_side_encryption_configuration.by_id", "rule.0.apply_server_side_encryption_by_default.0.sse_algorithm", "AES256"),
				),
			},
		},
	})
}

func TestAccDataSourceBucketServerSideEncryptionConfiguration_ByBucket(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCheckBucketServerSideEncryptionConfigurationDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_object_bucket" "main" {
					  name = "test-acc-scaleway-object-bucket-ds-sse-filter"
					  region = "fr-par"
					}

					resource "scaleway_object_bucket_server_side_encryption_configuration" "main" {
					  bucket = scaleway_object_bucket.main.name
					  rule {
						apply_server_side_encryption_by_default {
						  sse_algorithm = "AES256"
						}
					  }
					}

					data "scaleway_object_bucket_server_side_encryption_configuration" "by_bucket" {
					  bucket = scaleway_object_bucket.main.name
					  depends_on = [scaleway_object_bucket_server_side_encryption_configuration.main]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.scaleway_object_bucket_server_side_encryption_configuration.by_bucket", "bucket",
						"scaleway_object_bucket_server_side_encryption_configuration.main", "bucket"),
					resource.TestCheckResourceAttr("data.scaleway_object_bucket_server_side_encryption_configuration.by_bucket", "rule.0.apply_server_side_encryption_by_default.0.sse_algorithm", "AES256"),
				),
			},
		},
	})
}
