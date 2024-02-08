package scaleway

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const (
	LockResourcePrefix   = "tf-acc-test"
	lockResourceTestName = "scaleway_object_bucket_lock_configuration.test"
)

func TestAccScalewayObjectBucketLockConfiguration_Basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(LockResourcePrefix)
	resourceName := lockResourceTestName

	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        ErrorCheck(t, EndpointsID),
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayBucketLockConfigurationDestroy(tt),
			testAccCheckScalewayObjectBucketDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "test" {
						name = %[1]q
						region = %[2]q
						tags = {
							TestName = "TestAccSCW_LockConfig_basic"
						}

						object_lock_enabled = true
					}

					resource "scaleway_object_bucket_acl" "test" {
						bucket = scaleway_object_bucket.test.id
						acl = "public-read"
					}

					resource "scaleway_object_bucket_lock_configuration" "test" {
						bucket = scaleway_object_bucket.test.id
						rule {
							default_retention {
								mode = "GOVERNANCE"
								days = 1
							}
						}
					}
				`, rName, objectTestsMainRegion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLockConfigurationExists(tt, resourceName),
					testAccCheckScalewayObjectBucketExists(tt, "scaleway_object_bucket.test", true),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "scaleway_object_bucket.test", "name"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.default_retention.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.default_retention.0.mode", "GOVERNANCE"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.default_retention.0.days", "1"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "test" {
						name = %[1]q
						region = %[2]q
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
								years = 1
							}
						}
					}
				`, rName, objectTestsMainRegion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLockConfigurationExists(tt, resourceName),
					testAccCheckScalewayObjectBucketExists(tt, "scaleway_object_bucket.test", true),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "scaleway_object_bucket.test", "name"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.default_retention.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.default_retention.0.mode", "GOVERNANCE"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.default_retention.0.years", "1"),
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

func TestAccScalewayObjectBucketLockConfiguration_Update(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(LockResourcePrefix)
	resourceName := lockResourceTestName

	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        ErrorCheck(t, EndpointsID),
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayBucketLockConfigurationDestroy(tt),
			testAccCheckScalewayObjectBucketDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "test" {
						name = %[1]q
						region = %[2]q
						tags = {
							TestName = "TestAccSCW_LockConfig_update"
						}

						object_lock_enabled = true
					}

					resource "scaleway_object_bucket_acl" "test" {
						bucket = scaleway_object_bucket.test.id
						acl = "public-read"
					}

				  	resource "scaleway_object_bucket_lock_configuration" "test" {
						bucket = scaleway_object_bucket.test.id
						rule {
							default_retention {
								mode = "GOVERNANCE"
								days = 1
							}
						}
				  	}
				`, rName, objectTestsMainRegion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLockConfigurationExists(tt, resourceName),
					testAccCheckScalewayObjectBucketExists(tt, "scaleway_object_bucket.test", true),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "test" {
						name = %[1]q
						region = %[2]q
						tags = {
							TestName = "TestAccSCW_LockConfig_basic"
						}

						object_lock_enabled = true
					}

					resource "scaleway_object_bucket_acl" "test" {
						bucket = scaleway_object_bucket.test.id
						acl = "public-read"
					}

				  	resource "scaleway_object_bucket_lock_configuration" "test" {
						bucket = scaleway_object_bucket.test.id
						rule {
							default_retention {
								mode = "COMPLIANCE"
								days = 2
							}
						}
				  	}
				`, rName, objectTestsMainRegion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLockConfigurationExists(tt, resourceName),
					testAccCheckScalewayObjectBucketExists(tt, "scaleway_object_bucket.test", true),
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

func TestAccScalewayObjectBucketLockConfiguration_WithBucketName(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(LockResourcePrefix)
	resourceName := lockResourceTestName

	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        ErrorCheck(t, EndpointsID),
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayBucketLockConfigurationDestroy(tt),
			testAccCheckScalewayObjectBucketDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "test" {
						name = %[1]q
						region = %[2]q
						tags = {
							TestName = "TestAccSCW_LockConfig_WithBucketName"
						}

						object_lock_enabled = true
					}

					resource "scaleway_object_bucket_acl" "test" {
						bucket = scaleway_object_bucket.test.id
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
				`, rName, objectTestsMainRegion),
				ExpectError: regexp.MustCompile("NoSuchBucket: The specified bucket does not exist"),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "test" {
						name = %[1]q
						region = %[2]q
						tags = {
							TestName = "TestAccSCW_LockConfig_WithBucketName"
						}

						object_lock_enabled = true
					}

					resource "scaleway_object_bucket_acl" "test" {
						bucket = scaleway_object_bucket.test.id
						acl = "public-read"
					}

					resource "scaleway_object_bucket_lock_configuration" "test" {
						bucket = scaleway_object_bucket.test.name
						region = %[2]q
						rule {
							default_retention {
								mode = "GOVERNANCE"
								days = 1
							}
						}
					}
				`, rName, objectTestsMainRegion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLockConfigurationExists(tt, resourceName),
					testAccCheckScalewayObjectBucketExists(tt, "scaleway_object_bucket.test", true),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "scaleway_object_bucket.test", "name"),
				),
			},
		},
	})
}

func testAccCheckScalewayBucketLockConfigurationDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "scaleway_object_bucket_lock_configuration" {
				continue
			}

			regionalID := expandRegionalID(rs.Primary.ID)
			bucketRegion := regionalID.Region
			bucket := regionalID.ID
			conn, err := newS3ClientFromMeta(tt.Meta, bucketRegion.String())
			if err != nil {
				return err
			}

			input := &s3.GetObjectLockConfigurationInput{
				Bucket: aws.String(bucket),
			}

			output, err := conn.GetObjectLockConfiguration(input)

			if isS3Err(err, s3.ErrCodeNoSuchBucket, "") {
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
			return errors.New("resource not found")
		}

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource (%s) ID not set", resourceName)
		}

		regionalID := expandRegionalID(rs.Primary.ID)
		bucketRegion := regionalID.Region
		bucket := regionalID.ID
		conn, err := newS3ClientFromMeta(tt.Meta, bucketRegion.String())
		if err != nil {
			return err
		}

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
