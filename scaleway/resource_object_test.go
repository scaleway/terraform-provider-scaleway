package scaleway

import (
	"fmt"
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
		CheckDestroy:      testAccCheckScalewayObjectDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "base-01" {
						name = "%s"
						tags = {
							foo = "bar"
						}
					}
					
					resource scaleway_object "file" {
						bucket = scaleway_object_bucket.base-01.name
						key = "myfile"
					}
				`, bucketName),
				Check: resource.ComposeTestCheckFunc(
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
		CheckDestroy:      testAccCheckScalewayObjectDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "base-01" {
						name = "%s"
						tags = {
							foo = "bar"
						}
					}
					
					resource scaleway_object "file" {
						bucket = scaleway_object_bucket.base-01.name
						key = "myfile"
						hash = "1"
					}
				`, bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayObjectExists(tt, "scaleway_object.file"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "base-01" {
						name = "%s"
						tags = {
							foo = "bar"
						}
					}
					
					resource scaleway_object "file" {
						bucket = scaleway_object_bucket.base-01.name
						key = "myfile"
						hash = "2"
					}
				`, bucketName),
				Check: resource.ComposeTestCheckFunc(
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
		CheckDestroy:      testAccCheckScalewayObjectDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "base-01" {
						name = "%s"
						tags = {
							foo = "bar"
						}
					}
					
					resource scaleway_object "file" {
						bucket = scaleway_object_bucket.base-01.name
						key = "myfile"
					}
				`, bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayObjectExists(tt, "scaleway_object.file")),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "base-01" {
						name = "%s"
						tags = {
							foo = "bar"
						}
					}
					
					resource scaleway_object "file" {
						bucket = scaleway_object_bucket.base-01.name
						key = "myfile2"
					}
				`, bucketName),
				Check: resource.ComposeTestCheckFunc(
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
		CheckDestroy:      testAccCheckScalewayObjectDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "base-01" {
						name = "%s"
						tags = {
							foo = "bar"
						}
					}
					
					resource scaleway_object "file" {
						bucket = scaleway_object_bucket.base-01.name
						key = "myfile"

						storage_class = "ONEZONE_IA"
					}
				`, bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayObjectExists(tt, "scaleway_object.file"),
					resource.TestCheckResourceAttr("scaleway_object.file", "storage_class", "ONEZONE_IA"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "base-01" {
						name = "%s"
						tags = {
							foo = "bar"
						}
					}
					
					resource scaleway_object "file" {
						bucket = scaleway_object_bucket.base-01.name
						key = "myfile"

						storage_class = "STANDARD"
					}
				`, bucketName),
				Check: resource.ComposeTestCheckFunc(
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
		CheckDestroy:      testAccCheckScalewayObjectDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "base-01" {
						name = "%s"
						tags = {
							foo = "bar"
						}
					}
					
					resource scaleway_object "file" {
						bucket = scaleway_object_bucket.base-01.name
						key = "myfile"

						metadata = {
							key = "value"
						}
					}
				`, bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayObjectExists(tt, "scaleway_object.file"),
					resource.TestCheckResourceAttr("scaleway_object.file", "metadata.key", "value"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "base-01" {
						name = "%s"
						tags = {
							foo = "bar"
						}
					}
					
					resource scaleway_object "file" {
						bucket = scaleway_object_bucket.base-01.name
						key = "myfile"

						metadata = {
							key = "other_value"
							other_key = "VALUE"
						}
					}
				`, bucketName),
				Check: resource.ComposeTestCheckFunc(
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
		CheckDestroy:      testAccCheckScalewayObjectDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "base-01" {
						name = "%s"
					}
					
					resource scaleway_object "file" {
						bucket = scaleway_object_bucket.base-01.name
						key = "myfile"

						tags = {
							key = "value"
						}
					}
				`, bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayObjectExists(tt, "scaleway_object.file"),
					resource.TestCheckResourceAttr("scaleway_object.file", "tags.key", "value"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "base-01" {
						name = "%s"
					}
					
					resource scaleway_object "file" {
						bucket = scaleway_object_bucket.base-01.name
						key = "myfile"

						tags = {
							key = "other_value"
							other_key = "VALUE"
						}
					}
				`, bucketName),
				Check: resource.ComposeTestCheckFunc(
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
		CheckDestroy:      testAccCheckScalewayObjectDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "base-01" {
						name = "%s"
					}
					
					resource scaleway_object "file" {
						bucket = scaleway_object_bucket.base-01.name
						key = "myfile"

						visibility = "public-read"
					}
				`, bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayObjectExists(tt, "scaleway_object.file"),
					resource.TestCheckResourceAttr("scaleway_object.file", "visibility", "public-read"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "base-01" {
						name = "%s"
					}
					
					resource scaleway_object "file" {
						bucket = scaleway_object_bucket.base-01.name
						key = "myfile"

						visibility = "private"
					}
				`, bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayObjectExists(tt, "scaleway_object.file"),
					resource.TestCheckResourceAttr("scaleway_object.file", "visibility", "private"),
				),
			},
		},
	})
}

func TestAccScalewayObject_Import(t *testing.T) {
	if !*UpdateCassettes {
		t.Skip("Skipping ObjectStorage test as this kind of resource can't be deleted before 24h")
	}
	tt := NewTestTools(t)
	defer tt.Cleanup()
	bucketName := sdkacctest.RandomWithPrefix("test-acc-scaleway-object-visibility")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayObjectDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "base-01" {
						name = "%s"
					}
					
					resource scaleway_object "file" {
						bucket = scaleway_object_bucket.base-01.name
						key = "myfile"

						visibility = "public-read"
					}
				`, bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayObjectExists(tt, "scaleway_object.file"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "base-01" {
						name = "%s"
					}
					
					resource scaleway_object "file" {
						bucket = scaleway_object_bucket.base-01.name
						key = "myfile"

						visibility = "public-read"
					}

					resource scaleway_object "file_imported" {
						bucket = scaleway_object_bucket.base-01.name
						key = "myfile"
					}
				`, bucketName),
				ImportState:   true,
				ResourceName:  "scaleway_object.file_imported",
				ImportStateId: fmt.Sprintf("fr-par/%s/myfile", bucketName),
				Check: resource.ComposeTestCheckFunc(
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

func testAccCheckScalewayObjectExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs := state.RootModule().Resources[n]
		if rs == nil {
			return fmt.Errorf("resource not found")
		}
		bucketName := rs.Primary.Attributes["bucket"]
		key := rs.Primary.Attributes["key"]

		s3Client, err := newS3ClientFromMeta(tt.Meta)
		if err != nil {
			return err
		}

		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		_, err = s3Client.GetObject(&s3.GetObjectInput{
			Bucket: scw.StringPtr(bucketName),
			Key:    scw.StringPtr(key),
		})

		if err != nil {
			if isS3Err(err, s3.ErrCodeNoSuchBucket, "") {
				return fmt.Errorf("s3 object not found")
			}
			return err
		}
		return nil
	}
}

func testAccCheckScalewayObjectDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		s3Client, err := newS3ClientFromMeta(tt.Meta)
		if err != nil {
			return err
		}

		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway" {
				continue
			}

			bucketName := rs.Primary.Attributes["bucket"]
			key := rs.Primary.Attributes["key"]

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

			return fmt.Errorf("object should be deleted")
		}
		return nil
	}
}
