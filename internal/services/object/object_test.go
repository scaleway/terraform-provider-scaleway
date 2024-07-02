package object_test

import (
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/object"
	objectchecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/object/testfuncs"
)

// Service information constants
const (
	ServiceName      = "scw"       // Name of service.
	EndpointsID      = ServiceName // ID to look up a service endpoint with.
	fileContentStep2 = "This is a different content"
	fileContentStep1 = "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."
)

func TestAccObject_Basic(t *testing.T) {
	if !*acctest.UpdateCassettes {
		t.Skip("Skipping ObjectStorage test as this kind of resource can't be deleted before 24h")
	}
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	bucketName := sdkacctest.RandomWithPrefix("test-acc-scaleway-object-basic")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
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
					}
				`, bucketName, objectTestsMainRegion),
				Check: resource.ComposeTestCheckFunc(
					objectchecks.CheckBucketExists(tt, "scaleway_object_bucket.base-01", true),
					testAccCheckObjectExists(tt, "scaleway_object.file"),
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
						key = "myfile/foo"
					}
				`, bucketName, objectTestsMainRegion),
				Check: resource.ComposeTestCheckFunc(
					objectchecks.CheckBucketExists(tt, "scaleway_object_bucket.base-01", true),
					testAccCheckObjectExists(tt, "scaleway_object.file"),
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
						key = "myfile/foo/bar"
					}
				`, bucketName, objectTestsMainRegion),
				Check: resource.ComposeTestCheckFunc(
					objectchecks.CheckBucketExists(tt, "scaleway_object_bucket.base-01", true),
					testAccCheckObjectExists(tt, "scaleway_object.file"),
				),
			},
		},
	})
}

func TestAccObject_Hash(t *testing.T) {
	if !*acctest.UpdateCassettes {
		t.Skip("Skipping ObjectStorage test as this kind of resource can't be deleted before 24h")
	}
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	bucketName := sdkacctest.RandomWithPrefix("test-acc-scaleway-object-hash")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			objectchecks.IsObjectDestroyed(tt),
			objectchecks.IsBucketDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "base-01" {
						name = "%s"
						region = "%s"
						tags = {
							foo = "bar"
						}
					}
					
					resource scaleway_object "file" {
						bucket = scaleway_object_bucket.base-01.id
						key = "myfile"
						hash = "1"
					}
				`, bucketName, objectTestsMainRegion),
				Check: resource.ComposeTestCheckFunc(
					objectchecks.CheckBucketExists(tt, "scaleway_object_bucket.base-01", true),
					testAccCheckObjectExists(tt, "scaleway_object.file"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "base-01" {
						name = "%s"
						region = "%s"
						tags = {
							foo = "bar"
						}
					}
					
					resource scaleway_object "file" {
						bucket = scaleway_object_bucket.base-01.id
						key = "myfile"
						hash = "2"
					}
				`, bucketName, objectTestsMainRegion),
				Check: resource.ComposeTestCheckFunc(
					objectchecks.CheckBucketExists(tt, "scaleway_object_bucket.base-01", true),
					testAccCheckObjectExists(tt, "scaleway_object.file"),
				),
			},
		},
	})
}

func TestAccObject_Move(t *testing.T) {
	if !*acctest.UpdateCassettes {
		t.Skip("Skipping ObjectStorage test as this kind of resource can't be deleted before 24h")
	}
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	bucketName := sdkacctest.RandomWithPrefix("test-acc-scaleway-object-move")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			objectchecks.IsObjectDestroyed(tt),
			objectchecks.IsBucketDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "base-01" {
						name = "%s"
						region = "%s"
						tags = {
							foo = "bar"
						}
					}
					
					resource scaleway_object "file" {
						bucket = scaleway_object_bucket.base-01.id
						key = "myfile"
					}
				`, bucketName, objectTestsMainRegion),
				Check: resource.ComposeTestCheckFunc(
					objectchecks.CheckBucketExists(tt, "scaleway_object_bucket.base-01", true),
					testAccCheckObjectExists(tt, "scaleway_object.file")),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "base-01" {
						name = "%s"
						region = "%s"
						tags = {
							foo = "bar"
						}
					}
					
					resource scaleway_object "file" {
						bucket = scaleway_object_bucket.base-01.id
						key = "myfile2"
					}
				`, bucketName, objectTestsMainRegion),
				Check: resource.ComposeTestCheckFunc(
					objectchecks.CheckBucketExists(tt, "scaleway_object_bucket.base-01", true),
					testAccCheckObjectExists(tt, "scaleway_object.file")),
			},
		},
	})
}

func TestAccObject_StorageClass(t *testing.T) {
	if !*acctest.UpdateCassettes {
		t.Skip("Skipping ObjectStorage test as this kind of resource can't be deleted before 24h")
	}
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	bucketName := sdkacctest.RandomWithPrefix("test-acc-scaleway-object-storage-class")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			objectchecks.IsObjectDestroyed(tt),
			objectchecks.IsBucketDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "base-01" {
						name = "%s"
						region = "%s"
						tags = {
							foo = "bar"
						}
					}
					
					resource scaleway_object "file" {
						bucket = scaleway_object_bucket.base-01.id
						key = "myfile"

						storage_class = "ONEZONE_IA"
					}
				`, bucketName, objectTestsMainRegion),
				Check: resource.ComposeTestCheckFunc(
					objectchecks.CheckBucketExists(tt, "scaleway_object_bucket.base-01", true),
					testAccCheckObjectExists(tt, "scaleway_object.file"),
					resource.TestCheckResourceAttr("scaleway_object.file", "storage_class", "ONEZONE_IA"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "base-01" {
						name = "%s"
						region = "%s"
						tags = {
							foo = "bar"
						}
					}
					
					resource scaleway_object "file" {
						bucket = scaleway_object_bucket.base-01.id
						key = "myfile"

						storage_class = "STANDARD"
					}
				`, bucketName, objectTestsMainRegion),
				Check: resource.ComposeTestCheckFunc(
					objectchecks.CheckBucketExists(tt, "scaleway_object_bucket.base-01", true),
					testAccCheckObjectExists(tt, "scaleway_object.file"),
					resource.TestCheckResourceAttr("scaleway_object.file", "storage_class", "STANDARD"),
				),
			},
		},
	})
}

func TestAccObject_Metadata(t *testing.T) {
	if !*acctest.UpdateCassettes {
		t.Skip("Skipping ObjectStorage test as this kind of resource can't be deleted before 24h")
	}
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	bucketName := sdkacctest.RandomWithPrefix("test-acc-scaleway-object-metadata")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			objectchecks.IsObjectDestroyed(tt),
			objectchecks.IsBucketDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "base-01" {
						name = "%s"
						region = "%s"
						tags = {
							foo = "bar"
						}
					}
					
					resource scaleway_object "file" {
						bucket = scaleway_object_bucket.base-01.id
						key = "myfile"

						metadata = {
							key = "value"
						}
					}
				`, bucketName, objectTestsMainRegion),
				Check: resource.ComposeTestCheckFunc(
					objectchecks.CheckBucketExists(tt, "scaleway_object_bucket.base-01", true),
					testAccCheckObjectExists(tt, "scaleway_object.file"),
					resource.TestCheckResourceAttr("scaleway_object.file", "metadata.key", "value"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "base-01" {
						name = "%s"
						region = "%s"
						tags = {
							foo = "bar"
						}
					}
					
					resource scaleway_object "file" {
						bucket = scaleway_object_bucket.base-01.id
						key = "myfile"

						metadata = {
							key = "other_value"
							other_key = "VALUE"
						}
					}
				`, bucketName, objectTestsMainRegion),
				Check: resource.ComposeTestCheckFunc(
					objectchecks.CheckBucketExists(tt, "scaleway_object_bucket.base-01", true),
					testAccCheckObjectExists(tt, "scaleway_object.file"),
					resource.TestCheckResourceAttr("scaleway_object.file", "metadata.key", "other_value"),
					resource.TestCheckResourceAttr("scaleway_object.file", "metadata.other_key", "VALUE"),
				),
			},
		},
	})
}

func TestAccObject_Tags(t *testing.T) {
	if !*acctest.UpdateCassettes {
		t.Skip("Skipping ObjectStorage test as this kind of resource can't be deleted before 24h")
	}
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	bucketName := sdkacctest.RandomWithPrefix("test-acc-scaleway-object-tags")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			objectchecks.IsObjectDestroyed(tt),
			objectchecks.IsBucketDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "base-01" {
						name = "%s"
						region = "%s"
					}
					
					resource scaleway_object "file" {
						bucket = scaleway_object_bucket.base-01.id
						key = "myfile"

						tags = {
							key = "value"
						}
					}
				`, bucketName, objectTestsMainRegion),
				Check: resource.ComposeTestCheckFunc(
					objectchecks.CheckBucketExists(tt, "scaleway_object_bucket.base-01", true),
					testAccCheckObjectExists(tt, "scaleway_object.file"),
					resource.TestCheckResourceAttr("scaleway_object.file", "tags.key", "value"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "base-01" {
						name = "%s"
						region = "%s"
					}
					
					resource scaleway_object "file" {
						bucket = scaleway_object_bucket.base-01.id
						key = "myfile"

						tags = {
							key = "other_value"
							other_key = "VALUE"
						}
					}
				`, bucketName, objectTestsMainRegion),
				Check: resource.ComposeTestCheckFunc(
					objectchecks.CheckBucketExists(tt, "scaleway_object_bucket.base-01", true),
					testAccCheckObjectExists(tt, "scaleway_object.file"),
					resource.TestCheckResourceAttr("scaleway_object.file", "tags.key", "other_value"),
					resource.TestCheckResourceAttr("scaleway_object.file", "tags.other_key", "VALUE"),
				),
			},
		},
	})
}

func TestAccObject_Visibility(t *testing.T) {
	if !*acctest.UpdateCassettes {
		t.Skip("Skipping ObjectStorage test as this kind of resource can't be deleted before 24h")
	}
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	bucketName := sdkacctest.RandomWithPrefix("test-acc-scaleway-object-visibility")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			objectchecks.IsObjectDestroyed(tt),
			objectchecks.IsBucketDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "base-01" {
						name = "%s"
						region = "%s"
					}
					
					resource scaleway_object "file" {
						bucket = scaleway_object_bucket.base-01.id
						key = "myfile"

						visibility = "public-read"
					}
				`, bucketName, objectTestsMainRegion),
				Check: resource.ComposeTestCheckFunc(
					objectchecks.CheckBucketExists(tt, "scaleway_object_bucket.base-01", true),
					testAccCheckObjectExists(tt, "scaleway_object.file"),
					resource.TestCheckResourceAttr("scaleway_object.file", "visibility", "public-read"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "base-01" {
						name = "%s"
						region = "%s"
					}
					
					resource scaleway_object "file" {
						bucket = scaleway_object_bucket.base-01.id
						key = "myfile"

						visibility = "private"
					}
				`, bucketName, objectTestsMainRegion),
				Check: resource.ComposeTestCheckFunc(
					objectchecks.CheckBucketExists(tt, "scaleway_object_bucket.base-01", true),
					testAccCheckObjectExists(tt, "scaleway_object.file"),
					resource.TestCheckResourceAttr("scaleway_object.file", "visibility", "private"),
				),
			},
		},
	})
}

func TestAccObject_State(t *testing.T) {
	if !*acctest.UpdateCassettes {
		t.Skip("Skipping ObjectStorage test as this kind of resource can't be deleted before 24h")
	}
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	bucketName := sdkacctest.RandomWithPrefix("test-acc-scaleway-object-visibility")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			objectchecks.IsObjectDestroyed(tt),
			objectchecks.IsBucketDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "base-01" {
						name = "%s"
						region = "%s"
					}
					
					resource scaleway_object "file" {
						bucket = scaleway_object_bucket.base-01.id
						key = "myfile"

						visibility = "public-read"
					}
				`, bucketName, objectTestsMainRegion),
				Check: resource.ComposeTestCheckFunc(
					objectchecks.CheckBucketExists(tt, "scaleway_object_bucket.base-01", true),
					testAccCheckObjectExists(tt, "scaleway_object.file"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "base-01" {
						name = "%s"
						region = "%s"
					}
					
					resource scaleway_object "file" {
						bucket = scaleway_object_bucket.base-01.id
						key = "myfile"

						visibility = "public-read"
					}

					resource scaleway_object "file_imported" {
						bucket = scaleway_object_bucket.base-01.id
						key = "myfile"
					}
				`, bucketName, objectTestsMainRegion),
				ImportState:   true,
				ResourceName:  "scaleway_object.file_imported",
				ImportStateId: fmt.Sprintf("%s/%s/myfile", objectTestsMainRegion, bucketName),
				Check: resource.ComposeTestCheckFunc(
					objectchecks.CheckBucketExists(tt, "scaleway_object_bucket.base-01", true),
					testAccCheckObjectExists(tt, "scaleway_object.file"),
					testAccCheckObjectExists(tt, "scaleway_object.file_imported"),
					resource.TestCheckResourceAttrPair("scaleway_object.file_imported", "id", "scaleway_object.file", "id"),
					resource.TestCheckResourceAttrPair("scaleway_object.file_imported", "visibility", "scaleway_object.file", "visibility"),
					resource.TestCheckResourceAttrPair("scaleway_object.file_imported", "bucket", "scaleway_object.file", "bucket"),
					resource.TestCheckResourceAttrPair("scaleway_object.file_imported", "key", "scaleway_object.file", "key"),
				),
			},
		},
	})
}

func TestAccObject_ByContent(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	bucketName := sdkacctest.RandomWithPrefix("test-acc-scaleway-object-by-content")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			objectchecks.IsObjectDestroyed(tt),
			objectchecks.IsBucketDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "base-01" {
						name = "%s"
						region = "%s"
					}
					
					resource scaleway_object "by-content" {
						bucket = scaleway_object_bucket.base-01.id
						key = "test-by-content"
						content = "%s"
					}
				`, bucketName, objectTestsMainRegion, fileContentStep1),
				Check: resource.ComposeTestCheckFunc(
					objectchecks.CheckBucketExists(tt, "scaleway_object_bucket.base-01", true),
					testAccCheckObjectExists(tt, "scaleway_object.by-content"),
					resource.TestCheckResourceAttr("scaleway_object.by-content", "content", fileContentStep1),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "base-01" {
						name = "%s"
						region = "%s"
					}
					
					resource scaleway_object "by-content" {
						bucket = scaleway_object_bucket.base-01.id
						key = "test-by-content"
						content = "%s"
					}
				`, bucketName, objectTestsMainRegion, fileContentStep2),
				Check: resource.ComposeTestCheckFunc(
					objectchecks.CheckBucketExists(tt, "scaleway_object_bucket.base-01", true),
					testAccCheckObjectExists(tt, "scaleway_object.by-content"),
					resource.TestCheckResourceAttr("scaleway_object.by-content", "content", fileContentStep2),
				),
			},
		},
	})
}

func TestAccObject_ByContentBase64(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	bucketName := sdkacctest.RandomWithPrefix("test-acc-scaleway-object-by-content-base64")

	fileEncodedStep1 := base64.StdEncoding.EncodeToString([]byte(fileContentStep1))
	fileEncodedStep2 := base64.StdEncoding.EncodeToString([]byte(fileContentStep2))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			objectchecks.IsObjectDestroyed(tt),
			objectchecks.IsBucketDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "base-01" {
						name = "%s"
						region = "%s"
					}
					
					resource scaleway_object "by-content-base64" {
						bucket = scaleway_object_bucket.base-01.id
						key = "test-by-content-base64"
						content_base64 = base64encode("%s")
					}
				`, bucketName, objectTestsMainRegion, fileContentStep1),
				Check: resource.ComposeTestCheckFunc(
					objectchecks.CheckBucketExists(tt, "scaleway_object_bucket.base-01", true),
					testAccCheckObjectExists(tt, "scaleway_object.by-content-base64"),
					resource.TestCheckResourceAttr("scaleway_object.by-content-base64", "content_base64", fileEncodedStep1),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "base-01" {
						name = "%s"
						region = "%s"
					}
					
					resource scaleway_object "by-content-base64" {
						bucket = scaleway_object_bucket.base-01.id
						key = "test-by-content-base64"
						content_base64 = base64encode("%s")
					}
				`, bucketName, objectTestsMainRegion, fileContentStep2),
				Check: resource.ComposeTestCheckFunc(
					objectchecks.CheckBucketExists(tt, "scaleway_object_bucket.base-01", true),
					testAccCheckObjectExists(tt, "scaleway_object.by-content-base64"),
					resource.TestCheckResourceAttr("scaleway_object.by-content-base64", "content_base64", fileEncodedStep2),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "base-01" {
						name = "%s"
						region = "%s"
					}
					
					resource scaleway_object "by-content-base64" {
						bucket = scaleway_object_bucket.base-01.id
						key = "test-by-content-base64"
						content_base64 = "%s"
					}
				`, bucketName, objectTestsMainRegion, fileContentStep2),
				ExpectError: regexp.MustCompile("illegal base64 data at input byte 4"),
			},
		},
	})
}

func TestAccObject_SSECustomer(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	bucketName := sdkacctest.RandomWithPrefix("test-acc-scaleway-object-sse-customer")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			objectchecks.IsObjectDestroyed(tt),
			objectchecks.IsBucketDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "base-01" {
						name = "%s"
						region = "%s"
					}
					
					resource scaleway_object "sse-c-encrypted" {
						bucket = scaleway_object_bucket.base-01.id
						key = "test-sse-c-encrypted"
						content = "%s"
						sse_customer_key = "mY5up3r4w3s0meK3y"
					}
				`, bucketName, objectTestsMainRegion, fileContentStep1),
				Check: resource.ComposeTestCheckFunc(
					objectchecks.CheckBucketExists(tt, "scaleway_object_bucket.base-01", true),
					testAccCheckObjectExists(tt, "scaleway_object.sse-c-encrypted"),
					resource.TestCheckResourceAttr("scaleway_object.sse-c-encrypted", "content", fileContentStep1),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "base-01" {
						name = "%s"
						region = "%s"
					}
					
					resource scaleway_object "sse-c-encrypted" {
						bucket = scaleway_object_bucket.base-01.id
						key = "test-by-content"
						content = "%s"
						sse_customer_key = "mY5up3r4w3s0meK3y"
					}
				`, bucketName, objectTestsMainRegion, fileContentStep2),
				Check: resource.ComposeTestCheckFunc(
					objectchecks.CheckBucketExists(tt, "scaleway_object_bucket.base-01", true),
					testAccCheckObjectExists(tt, "scaleway_object.sse-c-encrypted"),
					resource.TestCheckResourceAttr("scaleway_object.sse-c-encrypted", "content", fileContentStep2),
				),
			},
		},
	})
}

func TestAccObject_WithBucketName(t *testing.T) {
	if !*acctest.UpdateCassettes {
		t.Skip("Skipping ObjectStorage test as this kind of resource can't be deleted before 24h")
	}
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	bucketName := sdkacctest.RandomWithPrefix("test-acc-scaleway-object-basic")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
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
						bucket = scaleway_object_bucket.base-01.name
						key = "myfile"
					}
				`, bucketName, objectTestsMainRegion),
				ExpectError: regexp.MustCompile("NoSuchBucket: The specified bucket does not exist"),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "base-01" {
						name = "%[1]s"
						region= "%[2]s"
						tags = {
							foo = "bar"
						}
					}
					
					resource scaleway_object "file" {
						bucket = scaleway_object_bucket.base-01.name
						region = "%[2]s"
						key = "myfile"
					}
				`, bucketName, objectTestsMainRegion),
				Check: resource.ComposeTestCheckFunc(
					objectchecks.CheckBucketExists(tt, "scaleway_object_bucket.base-01", true),
					testAccCheckObjectExists(tt, "scaleway_object.file"),
				),
			},
		},
	})
}

func testAccCheckObjectExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs := state.RootModule().Resources[n]
		if rs == nil {
			return errors.New("resource not found")
		}
		key := rs.Primary.Attributes["key"]

		regionalID := regional.ExpandID(rs.Primary.Attributes["bucket"])
		bucketRegion := regionalID.Region.String()
		bucketName := regionalID.ID

		s3Client, err := object.NewS3ClientFromMeta(tt.Meta, bucketRegion)
		if err != nil {
			return err
		}

		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no ID is set")
		}

		_, err = s3Client.GetObject(&s3.GetObjectInput{
			Bucket: scw.StringPtr(bucketName),
			Key:    scw.StringPtr(key),
		})
		if err != nil {
			if object.IsS3Err(err, s3.ErrCodeNoSuchBucket, "") {
				return errors.New("s3 object not found")
			}
			return err
		}
		return nil
	}
}
