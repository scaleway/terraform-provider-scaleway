package object_test

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	objectchecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/object/testfuncs"
)

// // Service information constants
const (
	ServiceName     = "scw"       // Name of service.
	EndpointsID     = ServiceName // ID to look up a service endpoint with.
	encryptionStr   = "1234567890abcdef1234567890abcdef"
	contentToEncypt = "Hello World"
)

func TestAccObject_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	bucketName := sdkacctest.RandomWithPrefix("test-acc-scaleway-object-basic")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: tt.ProviderFactories,
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
						key = "myfile/foo"
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
						key = "myfile/foo/bar"
						file   = "testfixture/empty.qcow2"
					}
				`, bucketName, objectTestsMainRegion),
				Check: resource.ComposeTestCheckFunc(
					objectchecks.CheckBucketExists(tt, "scaleway_object_bucket.base-01", true),
					objectchecks.IsObjectExists(tt, "scaleway_object.file"),
				),
			},
		},
	})
}

func TestAccObject_ContentType(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	bucketName := sdkacctest.RandomWithPrefix("test-acc-scaleway-object-content-type")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			objectchecks.IsObjectDestroyed(tt),
			objectchecks.IsBucketDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "scaleway_object_bucket" "main" {
					name = "%s"
					region = "%s"
				}

				resource "scaleway_object" "file" {
					bucket     = scaleway_object_bucket.main.id
					key        = "index.html"
					file       = "testfixture/index.html"
					visibility = "public-read"
					content_type = "text/html"
				}
					
				`, bucketName, objectTestsMainRegion),
				Check: resource.ComposeTestCheckFunc(
					objectchecks.CheckBucketExists(tt, "scaleway_object_bucket.main", true),
					objectchecks.IsObjectExists(tt, "scaleway_object.file"),
					resource.TestCheckResourceAttr("scaleway_object.file", "content_type", "text/html"),
				),
			},
		},
	})
}

func TestAccObject_Hash(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	bucketName := sdkacctest.RandomWithPrefix("test-acc-scaleway-object-hash")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: tt.ProviderFactories,
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
						file   = "testfixture/empty.qcow2"
						hash = "1"
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
						region = "%s"
						tags = {
							foo = "bar"
						}
					}

					resource scaleway_object "file" {
						bucket = scaleway_object_bucket.base-01.id
						key = "myfile"
						file   = "testfixture/empty.qcow2"
						hash = "2"
					}
				`, bucketName, objectTestsMainRegion),
				Check: resource.ComposeTestCheckFunc(
					objectchecks.CheckBucketExists(tt, "scaleway_object_bucket.base-01", true),
					objectchecks.IsObjectExists(tt, "scaleway_object.file"),
				),
			},
		},
	})
}

func TestAccObject_Move(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	bucketName := sdkacctest.RandomWithPrefix("test-acc-scaleway-object-move")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: tt.ProviderFactories,
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
						file   = "testfixture/empty.qcow2"
					}
				`, bucketName, objectTestsMainRegion),
				Check: resource.ComposeTestCheckFunc(
					objectchecks.CheckBucketExists(tt, "scaleway_object_bucket.base-01", true),
					objectchecks.IsObjectExists(tt, "scaleway_object.file")),
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
						file   = "testfixture/empty.qcow2"
					}
				`, bucketName, objectTestsMainRegion),
				Check: resource.ComposeTestCheckFunc(
					objectchecks.CheckBucketExists(tt, "scaleway_object_bucket.base-01", true),
					objectchecks.IsObjectExists(tt, "scaleway_object.file")),
			},
		},
	})
}

func TestAccObject_StorageClass(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	bucketName := sdkacctest.RandomWithPrefix("test-acc-scaleway-object-storage-class")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: tt.ProviderFactories,
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
						file   = "testfixture/empty.qcow2"

						storage_class = "ONEZONE_IA"
					}
				`, bucketName, objectTestsMainRegion),
				Check: resource.ComposeTestCheckFunc(
					objectchecks.CheckBucketExists(tt, "scaleway_object_bucket.base-01", true),
					objectchecks.IsObjectExists(tt, "scaleway_object.file"),
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
						file   = "testfixture/empty.qcow2"

						storage_class = "STANDARD"
					}
				`, bucketName, objectTestsMainRegion),
				Check: resource.ComposeTestCheckFunc(
					objectchecks.CheckBucketExists(tt, "scaleway_object_bucket.base-01", true),
					objectchecks.IsObjectExists(tt, "scaleway_object.file"),
					resource.TestCheckResourceAttr("scaleway_object.file", "storage_class", "STANDARD"),
				),
			},
		},
	})
}

func TestAccObject_Metadata(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	bucketName := sdkacctest.RandomWithPrefix("test-acc-scaleway-object-metadata")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: tt.ProviderFactories,
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
						file   = "testfixture/empty.qcow2"

						metadata = {
							key = "value"
						}
					}
				`, bucketName, objectTestsMainRegion),
				Check: resource.ComposeTestCheckFunc(
					objectchecks.CheckBucketExists(tt, "scaleway_object_bucket.base-01", true),
					objectchecks.IsObjectExists(tt, "scaleway_object.file"),
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
						file   = "testfixture/empty.qcow2"

						metadata = {
							key = "other_value"
							other_key = "VALUE"
						}
					}
				`, bucketName, objectTestsMainRegion),
				Check: resource.ComposeTestCheckFunc(
					objectchecks.CheckBucketExists(tt, "scaleway_object_bucket.base-01", true),
					objectchecks.IsObjectExists(tt, "scaleway_object.file"),
					resource.TestCheckResourceAttr("scaleway_object.file", "metadata.key", "other_value"),
					resource.TestCheckResourceAttr("scaleway_object.file", "metadata.other_key", "VALUE"),
				),
			},
		},
	})
}

func TestAccObject_Tags(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	bucketName := sdkacctest.RandomWithPrefix("test-acc-scaleway-object-tags")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: tt.ProviderFactories,
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
						file   = "testfixture/empty.qcow2"

						tags = {
							key = "value"
						}
					}
				`, bucketName, objectTestsMainRegion),
				Check: resource.ComposeTestCheckFunc(
					objectchecks.CheckBucketExists(tt, "scaleway_object_bucket.base-01", true),
					objectchecks.IsObjectExists(tt, "scaleway_object.file"),
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
						file   = "testfixture/empty.qcow2"

						tags = {
							key = "other_value"
							other_key = "VALUE"
						}
					}
				`, bucketName, objectTestsMainRegion),
				Check: resource.ComposeTestCheckFunc(
					objectchecks.CheckBucketExists(tt, "scaleway_object_bucket.base-01", true),
					objectchecks.IsObjectExists(tt, "scaleway_object.file"),
					resource.TestCheckResourceAttr("scaleway_object.file", "tags.key", "other_value"),
					resource.TestCheckResourceAttr("scaleway_object.file", "tags.other_key", "VALUE"),
				),
			},
		},
	})
}

func TestAccObject_Visibility(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	bucketName := sdkacctest.RandomWithPrefix("test-acc-scaleway-object-visibility")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: tt.ProviderFactories,
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
						file   = "testfixture/empty.qcow2"

						visibility = "public-read"
					}
				`, bucketName, objectTestsMainRegion),
				Check: resource.ComposeTestCheckFunc(
					objectchecks.CheckBucketExists(tt, "scaleway_object_bucket.base-01", true),
					objectchecks.IsObjectExists(tt, "scaleway_object.file"),
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
						file   = "testfixture/empty.qcow2"

						visibility = "private"
					}
				`, bucketName, objectTestsMainRegion),
				Check: resource.ComposeTestCheckFunc(
					objectchecks.CheckBucketExists(tt, "scaleway_object_bucket.base-01", true),
					objectchecks.IsObjectExists(tt, "scaleway_object.file"),
					resource.TestCheckResourceAttr("scaleway_object.file", "visibility", "private"),
				),
			},
		},
	})
}

func TestAccObject_State(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	bucketName := sdkacctest.RandomWithPrefix("test-acc-scaleway-object-visibility")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: tt.ProviderFactories,
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
						file   = "testfixture/empty.qcow2"

						visibility = "public-read"
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
						region = "%s"
					}

					resource scaleway_object "file" {
						bucket = scaleway_object_bucket.base-01.id
						key = "myfile"
						file   = "testfixture/empty.qcow2"

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
					objectchecks.IsObjectExists(tt, "scaleway_object.file"),
					objectchecks.IsObjectExists(tt, "scaleway_object.file_imported"),
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

	fileContentStep1 := "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."
	fileContentStep2 := "This is a different content"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: tt.ProviderFactories,
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
					objectchecks.IsObjectExists(tt, "scaleway_object.by-content"),
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
					objectchecks.IsObjectExists(tt, "scaleway_object.by-content"),
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

	fileContentStep1 := "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."
	fileContentStep2 := "This is a different content"
	fileEncodedStep1 := base64.StdEncoding.EncodeToString([]byte(fileContentStep1))
	fileEncodedStep2 := base64.StdEncoding.EncodeToString([]byte(fileContentStep2))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: tt.ProviderFactories,
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
					objectchecks.IsObjectExists(tt, "scaleway_object.by-content-base64"),
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
					objectchecks.IsObjectExists(tt, "scaleway_object.by-content-base64"),
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

func TestAccObject_WithBucketName(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	bucketName := sdkacctest.RandomWithPrefix("test-acc-scaleway-object-basic")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: tt.ProviderFactories,
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
						file   = "testfixture/empty.qcow2"
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
						file   = "testfixture/empty.qcow2"
					}
				`, bucketName, objectTestsMainRegion),
				Check: resource.ComposeTestCheckFunc(
					objectchecks.CheckBucketExists(tt, "scaleway_object_bucket.base-01", true),
					objectchecks.IsObjectExists(tt, "scaleway_object.file"),
				),
			},
		},
	})
}

func TestAccObject_Encryption(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	bucketName := sdkacctest.RandomWithPrefix("test-acc-scaleway-object-encryption")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: tt.ProviderFactories,
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

					resource scaleway_object "by-content" {
						bucket = scaleway_object_bucket.base-01.id
						key = "myfile/foo"
						content = "Hello World"
						sse_customer_key = "%s"
					}
				`, bucketName, objectTestsMainRegion, encryptionStr),
				Check: resource.ComposeTestCheckFunc(
					objectchecks.CheckBucketExists(tt, "scaleway_object_bucket.base-01", true),
					resource.TestCheckResourceAttr("scaleway_object.by-content", "content", "Hello World"),
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

					resource scaleway_object "by-content" {
						bucket = scaleway_object_bucket.base-01.id
						key = "myfile/foo/bar"
						content = "Hello World"
						sse_customer_key = "%s"
					}
				`, bucketName, objectTestsMainRegion, encryptionStr),
				Check: resource.ComposeTestCheckFunc(
					objectchecks.CheckBucketExists(tt, "scaleway_object_bucket.base-01", true),
					resource.TestCheckResourceAttr("scaleway_object.by-content", "content", "Hello World"),
				),
			},
		},
	})
}
