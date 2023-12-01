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
	ResourcePrefix   = "tf-acc-test"
	resourceTestName = "scaleway_object_bucket_website_configuration.test"
)

func TestAccScalewayObjectBucketWebsiteConfiguration_Basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(ResourcePrefix)
	resourceName := resourceTestName

	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        ErrorCheck(t, EndpointsID),
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayObjectBucketWebsiteConfigurationDestroy(tt),
			testAccCheckScalewayObjectBucketDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
			  		resource "scaleway_object_bucket" "test" {
						name = %[1]q
						region = %[2]q
						acl  = "public-read"
						tags = {
							TestName = "TestAccSCW_WebsiteConfig_basic"
						}
					}
				
				  	resource "scaleway_object_bucket_website_configuration" "test" {
						bucket = scaleway_object_bucket.test.name
						index_document {
						  suffix = "index.html"
						}
				  	}
				`, rName, objectTestsMainRegion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayObjectBucketExistsForceRegion(tt, "scaleway_object_bucket.test", true),
					testAccCheckScalewayObjectBucketWebsiteConfigurationExists(tt, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "scaleway_object_bucket.test", "name"),
					resource.TestCheckResourceAttr(resourceName, "index_document.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "index_document.0.suffix", "index.html"),
					resource.TestCheckResourceAttrSet(resourceName, "website_domain"),
					resource.TestCheckResourceAttrSet(resourceName, "website_endpoint"),
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

func TestAccScalewayObjectBucketWebsiteConfiguration_WithPolicy(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(ResourcePrefix)
	resourceName := resourceTestName

	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        ErrorCheck(t, EndpointsID),
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayObjectBucketWebsiteConfigurationDestroy(tt),
			testAccCheckScalewayObjectBucketDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
			  		resource "scaleway_object_bucket" "test" {
						name = %[1]q
						region = %[2]q
						acl  = "public-read"
						tags = {
							TestName = "TestAccSCW_WebsiteConfig_basic"
						}
					}

					resource "scaleway_object_bucket_policy" "main" {
						bucket = scaleway_object_bucket.test.name
						policy = jsonencode(
						{
							"Version" = "2012-10-17",
							"Id" = "MyPolicy",
							"Statement" = [
							{
							   "Sid" = "GrantToEveryone",
							   "Effect" = "Allow",
							   "Principal" = "*",
							   "Action" = [
								  "s3:GetObject"
							   ],
							   "Resource":[
								  "%[1]s/*"
							   ]
							}
							]
						})
					}
				
				  	resource "scaleway_object_bucket_website_configuration" "test" {
						bucket = scaleway_object_bucket.test.name
						index_document {
						  suffix = "index.html"
						}
				  	}
				`, rName, objectTestsMainRegion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayObjectBucketExistsForceRegion(tt, "scaleway_object_bucket.test", true),
					testAccCheckScalewayObjectBucketWebsiteConfigurationExists(tt, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "scaleway_object_bucket.test", "name"),
					resource.TestCheckResourceAttr(resourceName, "index_document.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "index_document.0.suffix", "index.html"),
					resource.TestCheckResourceAttrSet(resourceName, "website_domain"),
					resource.TestCheckResourceAttrSet(resourceName, "website_endpoint"),
				),
				ExpectNonEmptyPlan: !*UpdateCassettes,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccScalewayObjectBucketWebsiteConfiguration_Update(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(ResourcePrefix)
	resourceName := resourceTestName

	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        ErrorCheck(t, EndpointsID),
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayObjectBucketWebsiteConfigurationDestroy(tt),
			testAccCheckScalewayObjectBucketDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
			  		resource "scaleway_object_bucket" "test" {
						name = %[1]q
						region = %[2]q
						acl  = "public-read"
						tags = {
							TestName = "TestAccSCW_WebsiteConfig_basic"
						}
					}

				  	resource "scaleway_object_bucket_website_configuration" "test" {
						bucket = scaleway_object_bucket.test.name
						index_document {
						  suffix = "index.html"
						}
				  	}
				`, rName, objectTestsMainRegion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayObjectBucketExistsForceRegion(tt, "scaleway_object_bucket.test", true),
					testAccCheckScalewayObjectBucketWebsiteConfigurationExists(tt, resourceName),
				),
			},
			{
				Config: fmt.Sprintf(`
			  		resource "scaleway_object_bucket" "test" {
						name = %[1]q
						region = %[2]q
						acl  = "public-read"
						tags = {
							TestName = "TestAccSCW_WebsiteConfig_basic"
						}
					}

				  	resource "scaleway_object_bucket_website_configuration" "test" {
						bucket = scaleway_object_bucket.test.name
						index_document {
						  suffix = "index.html"
						}

						error_document {
							key = "error.html"
						}
				  	}
				`, rName, objectTestsMainRegion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayObjectBucketExistsForceRegion(tt, "scaleway_object_bucket.test", true),
					testAccCheckScalewayObjectBucketWebsiteConfigurationExists(tt, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "scaleway_object_bucket.test", "name"),
					resource.TestCheckResourceAttr(resourceName, "index_document.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "index_document.0.suffix", "index.html"),
					resource.TestCheckResourceAttr(resourceName, "error_document.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "error_document.0.key", "error.html"),
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

func testAccCheckScalewayObjectBucketWebsiteConfigurationDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn, err := newS3ClientFromMeta(tt.Meta)
		if err != nil {
			return err
		}

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "scaleway_object_bucket_website_configuration" {
				continue
			}

			bucket := expandID(rs.Primary.ID)

			input := &s3.GetBucketWebsiteInput{
				Bucket: aws.String(bucket),
			}

			output, err := conn.GetBucketWebsite(input)

			if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket, ErrCodeNoSuchWebsiteConfiguration) {
				continue
			}

			if err != nil {
				return fmt.Errorf("error getting object bucket website configuration (%s): %w", rs.Primary.ID, err)
			}

			if output != nil {
				return fmt.Errorf("object bucket website configuration (%s) still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccCheckScalewayObjectBucketWebsiteConfigurationExists(tt *TestTools, resourceName string) resource.TestCheckFunc {
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

		input := &s3.GetBucketWebsiteInput{
			Bucket: aws.String(bucket),
		}

		output, err := conn.GetBucketWebsite(input)
		if err != nil {
			return fmt.Errorf("error getting object bucket website configuration (%s): %w", rs.Primary.ID, err)
		}

		if output == nil {
			return fmt.Errorf("object bucket website configuration (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}
