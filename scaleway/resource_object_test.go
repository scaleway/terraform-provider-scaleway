package scaleway

import (
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func TestAccScalewayObject_Basic(t *testing.T) {
	if !*UpdateCassettes {
		t.Skip("Skipping ObjectStorage test as this kind of resource can't be deleted before 24h")
	}
	tt := NewTestTools(t)
	defer tt.Cleanup()
	bucketName := sdkacctest.RandomWithPrefix("test-acc-scaleway-object-basic")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayObjectDestroy(tt),
			testAccCheckScalewayObjectBucketDestroy(tt),
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
					testAccCheckScalewayObjectBucketExists(tt, "scaleway_object_bucket.base-01", true),
					testAccCheckScalewayObjectExists(tt, "scaleway_object.file"),
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
					testAccCheckScalewayObjectBucketExists(tt, "scaleway_object_bucket.base-01", true),
					testAccCheckScalewayObjectExists(tt, "scaleway_object.file"),
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
					testAccCheckScalewayObjectBucketExists(tt, "scaleway_object_bucket.base-01", true),
					testAccCheckScalewayObjectExists(tt, "scaleway_object.file"),
				),
			},
		},
	})
}

func TestAccScalewayObject_Hash(t *testing.T) {
	if !*UpdateCassettes {
		t.Skip("Skipping ObjectStorage test as this kind of resource can't be deleted before 24h")
	}
	tt := NewTestTools(t)
	defer tt.Cleanup()
	bucketName := sdkacctest.RandomWithPrefix("test-acc-scaleway-object-hash")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayObjectDestroy(tt),
			testAccCheckScalewayObjectBucketDestroy(tt),
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
					testAccCheckScalewayObjectBucketExists(tt, "scaleway_object_bucket.base-01", true),
					testAccCheckScalewayObjectExists(tt, "scaleway_object.file"),
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
					testAccCheckScalewayObjectBucketExists(tt, "scaleway_object_bucket.base-01", true),
					testAccCheckScalewayObjectExists(tt, "scaleway_object.file"),
				),
			},
		},
	})
}

func TestAccScalewayObject_Move(t *testing.T) {
	if !*UpdateCassettes {
		t.Skip("Skipping ObjectStorage test as this kind of resource can't be deleted before 24h")
	}
	tt := NewTestTools(t)
	defer tt.Cleanup()
	bucketName := sdkacctest.RandomWithPrefix("test-acc-scaleway-object-move")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayObjectDestroy(tt),
			testAccCheckScalewayObjectBucketDestroy(tt),
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
					testAccCheckScalewayObjectBucketExists(tt, "scaleway_object_bucket.base-01", true),
					testAccCheckScalewayObjectExists(tt, "scaleway_object.file")),
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
					testAccCheckScalewayObjectBucketExists(tt, "scaleway_object_bucket.base-01", true),
					testAccCheckScalewayObjectExists(tt, "scaleway_object.file")),
			},
		},
	})
}

func TestAccScalewayObject_StorageClass(t *testing.T) {
	if !*UpdateCassettes {
		t.Skip("Skipping ObjectStorage test as this kind of resource can't be deleted before 24h")
	}
	tt := NewTestTools(t)
	defer tt.Cleanup()
	bucketName := sdkacctest.RandomWithPrefix("test-acc-scaleway-object-storage-class")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayObjectDestroy(tt),
			testAccCheckScalewayObjectBucketDestroy(tt),
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
					testAccCheckScalewayObjectBucketExists(tt, "scaleway_object_bucket.base-01", true),
					testAccCheckScalewayObjectExists(tt, "scaleway_object.file"),
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
					testAccCheckScalewayObjectBucketExists(tt, "scaleway_object_bucket.base-01", true),
					testAccCheckScalewayObjectExists(tt, "scaleway_object.file"),
					resource.TestCheckResourceAttr("scaleway_object.file", "storage_class", "STANDARD"),
				),
			},
		},
	})
}

func TestAccScalewayObject_Metadata(t *testing.T) {
	if !*UpdateCassettes {
		t.Skip("Skipping ObjectStorage test as this kind of resource can't be deleted before 24h")
	}
	tt := NewTestTools(t)
	defer tt.Cleanup()
	bucketName := sdkacctest.RandomWithPrefix("test-acc-scaleway-object-metadata")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayObjectDestroy(tt),
			testAccCheckScalewayObjectBucketDestroy(tt),
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
					testAccCheckScalewayObjectBucketExists(tt, "scaleway_object_bucket.base-01", true),
					testAccCheckScalewayObjectExists(tt, "scaleway_object.file"),
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
					testAccCheckScalewayObjectBucketExists(tt, "scaleway_object_bucket.base-01", true),
					testAccCheckScalewayObjectExists(tt, "scaleway_object.file"),
					resource.TestCheckResourceAttr("scaleway_object.file", "metadata.key", "other_value"),
					resource.TestCheckResourceAttr("scaleway_object.file", "metadata.other_key", "VALUE"),
				),
			},
		},
	})
}

func TestAccScalewayObject_Tags(t *testing.T) {
	if !*UpdateCassettes {
		t.Skip("Skipping ObjectStorage test as this kind of resource can't be deleted before 24h")
	}
	tt := NewTestTools(t)
	defer tt.Cleanup()
	bucketName := sdkacctest.RandomWithPrefix("test-acc-scaleway-object-tags")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayObjectDestroy(tt),
			testAccCheckScalewayObjectBucketDestroy(tt),
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
					testAccCheckScalewayObjectBucketExists(tt, "scaleway_object_bucket.base-01", true),
					testAccCheckScalewayObjectExists(tt, "scaleway_object.file"),
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
					testAccCheckScalewayObjectBucketExists(tt, "scaleway_object_bucket.base-01", true),
					testAccCheckScalewayObjectExists(tt, "scaleway_object.file"),
					resource.TestCheckResourceAttr("scaleway_object.file", "tags.key", "other_value"),
					resource.TestCheckResourceAttr("scaleway_object.file", "tags.other_key", "VALUE"),
				),
			},
		},
	})
}

func TestAccScalewayObject_Visibility(t *testing.T) {
	if !*UpdateCassettes {
		t.Skip("Skipping ObjectStorage test as this kind of resource can't be deleted before 24h")
	}
	tt := NewTestTools(t)
	defer tt.Cleanup()
	bucketName := sdkacctest.RandomWithPrefix("test-acc-scaleway-object-visibility")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayObjectDestroy(tt),
			testAccCheckScalewayObjectBucketDestroy(tt),
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
					testAccCheckScalewayObjectBucketExists(tt, "scaleway_object_bucket.base-01", true),
					testAccCheckScalewayObjectExists(tt, "scaleway_object.file"),
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
					testAccCheckScalewayObjectBucketExists(tt, "scaleway_object_bucket.base-01", true),
					testAccCheckScalewayObjectExists(tt, "scaleway_object.file"),
					resource.TestCheckResourceAttr("scaleway_object.file", "visibility", "private"),
				),
			},
		},
	})
}

func TestAccScalewayObject_State(t *testing.T) {
	if !*UpdateCassettes {
		t.Skip("Skipping ObjectStorage test as this kind of resource can't be deleted before 24h")
	}
	tt := NewTestTools(t)
	defer tt.Cleanup()
	bucketName := sdkacctest.RandomWithPrefix("test-acc-scaleway-object-visibility")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayObjectDestroy(tt),
			testAccCheckScalewayObjectBucketDestroy(tt),
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
					testAccCheckScalewayObjectBucketExists(tt, "scaleway_object_bucket.base-01", true),
					testAccCheckScalewayObjectExists(tt, "scaleway_object.file"),
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
					testAccCheckScalewayObjectBucketExists(tt, "scaleway_object_bucket.base-01", true),
					testAccCheckScalewayObjectExists(tt, "scaleway_object.file"),
					testAccCheckScalewayObjectExists(tt, "scaleway_object.file_imported"),
					resource.TestCheckResourceAttrPair("scaleway_object.file_imported", "id", "scaleway_object.file", "id"),
					resource.TestCheckResourceAttrPair("scaleway_object.file_imported", "visibility", "scaleway_object.file", "visibility"),
					resource.TestCheckResourceAttrPair("scaleway_object.file_imported", "bucket", "scaleway_object.file", "bucket"),
					resource.TestCheckResourceAttrPair("scaleway_object.file_imported", "key", "scaleway_object.file", "key"),
				),
			},
		},
	})
}

func TestAccScalewayObject_ByContent(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	bucketName := sdkacctest.RandomWithPrefix("test-acc-scaleway-object-by-content")

	fileContentStep1 := "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."
	fileContentStep2 := "This is a different content"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayObjectDestroy(tt),
			testAccCheckScalewayObjectBucketDestroy(tt),
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
					testAccCheckScalewayObjectBucketExists(tt, "scaleway_object_bucket.base-01", true),
					testAccCheckScalewayObjectExists(tt, "scaleway_object.by-content"),
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
					testAccCheckScalewayObjectBucketExists(tt, "scaleway_object_bucket.base-01", true),
					testAccCheckScalewayObjectExists(tt, "scaleway_object.by-content"),
					resource.TestCheckResourceAttr("scaleway_object.by-content", "content", fileContentStep2),
				),
			},
		},
	})
}

func TestAccScalewayObject_ByContentBase64(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	bucketName := sdkacctest.RandomWithPrefix("test-acc-scaleway-object-by-content-base64")

	fileContentStep1 := "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."
	fileContentStep2 := "This is a different content"
	fileEncodedStep1 := base64.StdEncoding.EncodeToString([]byte(fileContentStep1))
	fileEncodedStep2 := base64.StdEncoding.EncodeToString([]byte(fileContentStep2))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayObjectDestroy(tt),
			testAccCheckScalewayObjectBucketDestroy(tt),
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
					testAccCheckScalewayObjectBucketExists(tt, "scaleway_object_bucket.base-01", true),
					testAccCheckScalewayObjectExists(tt, "scaleway_object.by-content-base64"),
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
					testAccCheckScalewayObjectBucketExists(tt, "scaleway_object_bucket.base-01", true),
					testAccCheckScalewayObjectExists(tt, "scaleway_object.by-content-base64"),
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

func TestAccScalewayObject_WithBucketName(t *testing.T) {
	if !*UpdateCassettes {
		t.Skip("Skipping ObjectStorage test as this kind of resource can't be deleted before 24h")
	}
	tt := NewTestTools(t)
	defer tt.Cleanup()
	bucketName := sdkacctest.RandomWithPrefix("test-acc-scaleway-object-basic")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayObjectDestroy(tt),
			testAccCheckScalewayObjectBucketDestroy(tt),
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
					testAccCheckScalewayObjectBucketExists(tt, "scaleway_object_bucket.base-01", true),
					testAccCheckScalewayObjectExists(tt, "scaleway_object.file"),
				),
			},
		},
	})
}

func testAccCheckScalewayObjectExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs := state.RootModule().Resources[n]
		if rs == nil {
			return errors.New("resource not found")
		}
		key := rs.Primary.Attributes["key"]

		regionalID := expandRegionalID(rs.Primary.Attributes["bucket"])
		bucketRegion := regionalID.Region.String()
		bucketName := regionalID.ID

		s3Client, err := newS3ClientFromMeta(tt.Meta, bucketRegion)
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
			if isS3Err(err, s3.ErrCodeNoSuchBucket, "") {
				return errors.New("s3 object not found")
			}
			return err
		}
		return nil
	}
}

func testAccCheckScalewayObjectDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway" {
				continue
			}

			regionalID := expandRegionalID(rs.Primary.Attributes["bucket"])
			bucketRegion := regionalID.Region.String()
			bucketName := regionalID.ID
			key := rs.Primary.Attributes["key"]

			s3Client, err := newS3ClientFromMeta(tt.Meta, bucketRegion)
			if err != nil {
				return err
			}

			_, err = s3Client.GetObject(&s3.GetObjectInput{
				Bucket: scw.StringPtr(bucketName),
				Key:    scw.StringPtr(key),
			})
			if err != nil {
				if s3err, ok := err.(awserr.Error); ok && s3err.Code() == s3.ErrCodeNoSuchBucket {
					// bucket doesn't exist
					continue
				}
				return fmt.Errorf("couldn't get object to verify if it stil exists: %s", err)
			}

			return errors.New("object should be deleted")
		}
		return nil
	}
}
