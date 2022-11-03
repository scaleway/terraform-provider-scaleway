package scaleway

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const (
	LockResourcePrefix   = "tf-acc-test"
	lockResourceTestName = "scaleway_object_bucket_lock_configuration.test"
)

func TestAccObjectBucketLockConfiguration_basic(t *testing.T) {
	if !*UpdateCassettes {
		t.Skip("Skipping ObjectStorage test as this kind of resource can't be deleted before 24h")
	}
	rName := sdkacctest.RandomWithPrefix(LockResourcePrefix)
	resourceName := lockResourceTestName

	tt := NewTestTools(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        ErrorCheck(t, EndpointsID),
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckBucketLockConfigurationDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "test" {
						name = %[1]q
						tags = {
							TestName = "TestAccSCW_LockConfig_basic"
						}

						object_lock_enabled = true
					}

					resource "scaleway_object_bucket_acl" "test" {
						bucket = scaleway_object_bucket.test.name
						acl = "public-read"
					}

					resource "scaleway_object_bucket_lock_configuration" "test" {
						bucket = scaleway_object_bucket.test.name
						rule {
							default_retention {
								mode = "GOVERNANCE"
								days = 1
							}
						}
					}
				`, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLockConfigurationExists(tt, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "scaleway_object_bucket.test", "name"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.default_retention.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.default_retention.0.mode", "GOVERNANCE"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.default_retention.0.days", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccObjectBucketLockConfiguration_update(t *testing.T) {
	if !*UpdateCassettes {
		t.Skip("Skipping ObjectStorage test as this kind of resource can't be deleted before 24h")
	}
	rName := sdkacctest.RandomWithPrefix(LockResourcePrefix)
	resourceName := lockResourceTestName

	tt := NewTestTools(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        ErrorCheck(t, EndpointsID),
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckBucketLockConfigurationDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "test" {
						name = %[1]q
						tags = {
							TestName = "TestAccSCW_LockConfig_basic"
						}

						object_lock_enabled = true
					}

					resource "scaleway_object_bucket_acl" "test" {
						bucket = scaleway_object_bucket.test.name
						acl = "public-read"
					}

				  	resource "scaleway_object_bucket_lock_configuration" "test" {
						bucket = scaleway_object_bucket.test.name
						rule {
							default_retention {
								mode = "GOVERNANCE"
								days = 1
							}
						}
				  	}
				`, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLockConfigurationExists(tt, resourceName),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "test" {
						name = %[1]q
						tags = {
							TestName = "TestAccSCW_LockConfig_basic"
						}

						object_lock_enabled = true
					}

					resource "scaleway_object_bucket_acl" "test" {
						bucket = scaleway_object_bucket.test.name
						acl = "public-read"
					}

				  	resource "scaleway_object_bucket_lock_configuration" "test" {
						bucket = scaleway_object_bucket.test.name
						rule {
							default_retention {
								mode = "COMPLIANCE"
								days = 2
							}
						}
				  	}
				`, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLockConfigurationExists(tt, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "scaleway_object_bucket.test", "name"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.default_retention.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.default_retention.0.mode", "COMPLIANCE"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.default_retention.0.days", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckBucketLockConfigurationDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn, err := newS3ClientFromMeta(tt.Meta)
		if err != nil {
			return err
		}

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "scaleway_object_bucket_lock_configuration" {
				continue
			}

			bucket := expandID(rs.Primary.ID)

			input := &s3.GetObjectLockConfigurationInput{
				Bucket: aws.String(bucket),
			}

			output, err := conn.GetObjectLockConfiguration(input)

			if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
				continue
			}

			if err != nil {
				return fmt.Errorf("error getting object bucket lock configuration (%s): %w", rs.Primary.ID, err)
			}

			if output != nil {
				return fmt.Errorf("object bucket lock configuration (%s) still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccCheckBucketLockConfigurationExists(tt *TestTools, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[resourceName]
		if rs == nil {
			return fmt.Errorf("resource not found")
		}

		conn, err := newS3ClientFromMeta(tt.Meta)
		if err != nil {
			return err
		}

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource (%s) ID not set", resourceName)
		}

		bucket := expandID(rs.Primary.ID)

		input := &s3.GetObjectLockConfigurationInput{
			Bucket: aws.String(bucket),
		}

		output, err := conn.GetObjectLockConfiguration(input)
		if err != nil {
			return fmt.Errorf("error getting object bucket lock configuration (%s): %w", rs.Primary.ID, err)
		}

		if output == nil {
			return fmt.Errorf("object bucket lock configuration (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}
