package object_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	objectchecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/object/testfuncs"
)

func TestAccDataSourceObject_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	bucketName := sdkacctest.RandomWithPrefix("test-acc-scaleway-data-source-object-basic")
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			objectchecks.IsObjectDestroyed(tt),
			objectchecks.IsBucketDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "base-01" {
						name = "%s"
						region= "%s"
						tags = {
							foo = "bar"
						}
					}

					resource scaleway_object "file" {
						bucket = scaleway_object_bucket.base-01.id
						key = "myfile"
						file   = "testfixture/empty.qcow2"
					}

				`, bucketName, objectTestsMainRegion),
				Check: resource.ComposeTestCheckFunc(
					objectchecks.CheckBucketExists(tt, "scaleway_object_bucket.base-01", true),
					objectchecks.IsObjectExists(tt, "scaleway_object.file"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "base-01" {
						name = "%s"
						region= "%s"
						tags = {
							foo = "bar"
						}
					}

					resource scaleway_object "file" {
						bucket = scaleway_object_bucket.base-01.id
						key = "myfile"
						file   = "testfixture/empty.qcow2"
					}

					data scaleway_object "by-key" {
						key = "myfile"
						bucket = scaleway_object_bucket.base-01.id
					}
				`, bucketName, objectTestsMainRegion),
				Check: resource.ComposeTestCheckFunc(
					objectchecks.CheckBucketExists(tt, "scaleway_object_bucket.base-01", true),
					resource.TestCheckResourceAttr("data.scaleway_object.by-key", "key", "myfile"),
					resource.TestCheckResourceAttrSet("data.scaleway_object.by-key", "id"),
				),
			},
		},
	})
}

func TestAccDataSourceObject_Encrypted(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	bucketName := sdkacctest.RandomWithPrefix("test-acc-scaleway-ds-obj-encrypted")
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			objectchecks.IsObjectDestroyed(tt),
			objectchecks.IsBucketDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "base-01" {
						name = "%s"
						region= "%s"
						tags = {
							foo = "bar"
						}
					}

					resource scaleway_object "file" {
						bucket = scaleway_object_bucket.base-01.id
						key = "myfile"
						content = "Hello World"
						sse_customer_key = "%s"
					}

				`, bucketName, objectTestsMainRegion, encryptionStr),
				Check: resource.ComposeTestCheckFunc(
					objectchecks.CheckBucketExists(tt, "scaleway_object_bucket.base-01", true),
					objectchecks.IsObjectExists(tt, "scaleway_object.file"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "base-01" {
						name = "%s"
						region= "%s"
						tags = {
							foo = "bar"
						}
					}

					resource scaleway_object "file" {
						bucket = scaleway_object_bucket.base-01.id
						key = "myfile"
						content = "Hello World"
						sse_customer_key = "%s"
					}

					data scaleway_object "by-key" {
						key = "myfile"
						bucket = scaleway_object_bucket.base-01.id
						sse_customer_key = "%s"
					}
				`, bucketName, objectTestsMainRegion, encryptionStr, encryptionStr),
				Check: resource.ComposeTestCheckFunc(
					objectchecks.CheckBucketExists(tt, "scaleway_object_bucket.base-01", true),
					resource.TestCheckResourceAttr("data.scaleway_object.by-key", "key", "myfile"),
					resource.TestCheckResourceAttrSet("data.scaleway_object.by-key", "id"),
				),
			},
		},
	})
}

func TestAccDataSourceObject_EncryptedWO(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	bucketName := sdkacctest.RandomWithPrefix("test-acc-scaleway-ds-obj-encryptedwo")
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			objectchecks.IsObjectDestroyed(tt),
			objectchecks.IsBucketDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "base-01" {
						name = "%s"
						region= "%s"
						tags = {
							foo = "bar"
						}
					}

					resource scaleway_object "file" {
						bucket = scaleway_object_bucket.base-01.id
						key = "myfile"
						content = "Hello World"
						sse_customer_key_wo = "%s"
						sse_customer_key_wo_version = 1
					}

				`, bucketName, objectTestsMainRegion, encryptionStr),
				Check: resource.ComposeTestCheckFunc(
					objectchecks.CheckBucketExists(tt, "scaleway_object_bucket.base-01", true),
					objectchecks.IsObjectExists(tt, "scaleway_object.file"),
				),
			},
			// The only way to get an encrypted object is to provide the sse_customer_key. For a sse_customer_key_wo,
			// datasources cannot have Write Only attributes, so we have to pass it as a regular sse_customer_key.
			// This is not ideal, as the key is then set in the state, making the Write Only useless...
			// Querying objects encrypted with a sse_customer_key_wo is discouraged.
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "base-01" {
						name = "%s"
						region= "%s"
						tags = {
							foo = "bar"
						}
					}

					resource scaleway_object "file" {
						bucket = scaleway_object_bucket.base-01.id
						key = "myfile"
						content = "Hello World"
						sse_customer_key_wo = "%s"
						sse_customer_key_wo_version = 1
					}

					data scaleway_object "by-key" {
						key = "myfile"
						bucket = scaleway_object_bucket.base-01.id
						sse_customer_key = "%s"
					}
				`, bucketName, objectTestsMainRegion, encryptionStr, encryptionStr),
				Check: resource.ComposeTestCheckFunc(
					objectchecks.CheckBucketExists(tt, "scaleway_object_bucket.base-01", true),
					objectchecks.IsObjectExists(tt, "scaleway_object.file"),
				),
			},
		},
	})
}
