package scaleway

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func init() {
	resource.AddTestSweepers("scaleway_object_bucket", &resource.Sweeper{
		Name: "scaleway_object_bucket",
		F:    testSweepStorageObjectBucket,
	})
}

func TestAccScalewayObjectBucket_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	testBucketName := "test-acc-scaleway-object-bucket-basic"
	testBucketNameAms := testBucketName + "-ams"
	testBucketNamePar := testBucketName + "-par"
	testBucketACL := "private"
	testBucketUpdatedACL := "public-read"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayObjectBucketDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "base" {
						name = "%s"
						tags = {
							foo = "bar"
						}
					}

					resource "scaleway_object_bucket" "ams-bucket" {
						name = "%s"
						region = "nl-ams"
					}

					resource "scaleway_object_bucket" "par-bucket" {
						name = "%s"
						region = "fr-par"
					}
				`, testBucketName, testBucketNameAms, testBucketNamePar),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_object_bucket.base", "name", testBucketName),
					resource.TestCheckResourceAttr("scaleway_object_bucket.base", "acl", testBucketACL),
					resource.TestCheckResourceAttr("scaleway_object_bucket.base", "tags.%", "1"),
					resource.TestCheckResourceAttr("scaleway_object_bucket.base", "tags.foo", "bar"),
					resource.TestCheckResourceAttr("scaleway_object_bucket.base", "endpoint", fmt.Sprintf("https://%s.s3.%s.scw.cloud", testBucketName, "fr-par")),

					resource.TestCheckResourceAttr("scaleway_object_bucket.ams-bucket", "name", testBucketNameAms),
					resource.TestCheckResourceAttr("scaleway_object_bucket.ams-bucket", "endpoint", fmt.Sprintf("https://%s.s3.%s.scw.cloud", testBucketNameAms, "nl-ams")),

					resource.TestCheckResourceAttr("scaleway_object_bucket.par-bucket", "name", testBucketNamePar),
					resource.TestCheckResourceAttr("scaleway_object_bucket.par-bucket", "endpoint", fmt.Sprintf("https://%s.s3.%s.scw.cloud", testBucketNamePar, "fr-par")),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "base" {
						name = "%s"
						acl = "%s"
					}
				`, testBucketName, testBucketUpdatedACL),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_object_bucket.base", "name", testBucketName),
					resource.TestCheckResourceAttr("scaleway_object_bucket.base", "acl", testBucketUpdatedACL),
					resource.TestCheckResourceAttr("scaleway_object_bucket.base", "tags.%", "0"),
				),
			},
		},
	})
}

func testAccCheckScalewayObjectBucketDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		s3Client, err := newS3ClientFromMeta(tt.Meta)
		if err != nil {
			return err
		}

		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway" {
				continue
			}

			bucketName := rs.Primary.ID

			_, err := s3Client.ListObjects(&s3.ListObjectsInput{
				Bucket: &bucketName,
			})
			if err != nil {
				if s3err, ok := err.(awserr.Error); ok && s3err.Code() == s3.ErrCodeNoSuchBucket {
					// bucket doesn't exist
					continue
				}
				return fmt.Errorf("couldn't get bucket to verify if it stil exists: %s", err)
			}

			return fmt.Errorf("bucket should be deleted")
		}
		return nil
	}
}

func testSweepStorageObjectBucket(_ string) error {
	return sweepRegions([]scw.Region{scw.RegionFrPar, scw.RegionNlAms, scw.RegionPlWaw}, func(scwClient *scw.Client, region scw.Region) error {
		s3client, err := sharedS3ClientForRegion(region)
		if err != nil {
			return fmt.Errorf("error getting client: %s", err)
		}

		listBucketResponse, err := s3client.ListBuckets(&s3.ListBucketsInput{})
		if err != nil {
			return fmt.Errorf("couldn't list buckets: %s", err)
		}

		for _, bucket := range listBucketResponse.Buckets {
			l.Debugf("Deleting %q bucket", *bucket.Name)
			if strings.HasPrefix(*bucket.Name, "terraform-test") {
				_, err := s3client.DeleteBucket(&s3.DeleteBucketInput{
					Bucket: bucket.Name,
				})
				if err != nil {
					return fmt.Errorf("error deleting bucket in Sweeper: %s", err)
				}
			}
		}

		return nil
	})
}

func TestAccScalewayObjectBucket_Cors_Update(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	bucketName := "test-acc-scaleway-object-bucket-cors-update-000"
	const resourceName = "scaleway_object_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayObjectBucketDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "bucket" {
						name = %[1]q
						cors_rule {
							allowed_headers = ["*"]
							allowed_methods = ["PUT", "POST"]
							allowed_origins = ["https://www.example.com"]
							expose_headers  = ["x-amz-server-side-encryption", "ETag"]
							max_age_seconds = 3000
						}
					}`, bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayObjectBucketExists(tt, resourceName),
					testAccCheckScalewayObjectBucketCors(tt,
						resourceName,
						[]*s3.CORSRule{
							{
								AllowedHeaders: []*string{scw.StringPtr("*")},
								AllowedMethods: []*string{scw.StringPtr("PUT"), scw.StringPtr("POST")},
								AllowedOrigins: []*string{scw.StringPtr("https://www.example.com")},
								ExposeHeaders:  []*string{scw.StringPtr("x-amz-server-side-encryption"), scw.StringPtr("ETag")},
								MaxAgeSeconds:  scw.Int64Ptr(3000),
							},
						},
					),
					func(state *terraform.State) error {
						rs, ok := state.RootModule().Resources[resourceName]
						if !ok {
							return fmt.Errorf("not found: %s", resourceName)
						}

						s3Client, err := newS3ClientFromMeta(tt.Meta)
						if err != nil {
							return err
						}
						_, err = s3Client.PutBucketCors(&s3.PutBucketCorsInput{
							Bucket: scw.StringPtr(rs.Primary.Attributes["name"]),
							CORSConfiguration: &s3.CORSConfiguration{
								CORSRules: []*s3.CORSRule{
									{
										AllowedHeaders: []*string{scw.StringPtr("*")},
										AllowedMethods: []*string{scw.StringPtr("GET")},
										AllowedOrigins: []*string{scw.StringPtr("https://www.example.com")},
									},
								},
							},
						})
						if err != nil && !isS3Err(err, "NoSuchCORSConfiguration", "") {
							return err
						}
						return nil
					},
				),
				ExpectNonEmptyPlan: true,
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "bucket" {
						name = %[1]q
						cors_rule {
							allowed_headers = ["*"]
							allowed_methods = ["PUT", "POST"]
							allowed_origins = ["https://www.example.com"]
							expose_headers  = ["x-amz-server-side-encryption", "ETag"]
							max_age_seconds = 3000
						}
					}`, bucketName), Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayObjectBucketExists(tt, resourceName),
					testAccCheckScalewayObjectBucketCors(tt,
						resourceName,
						[]*s3.CORSRule{
							{
								AllowedHeaders: []*string{scw.StringPtr("*")},
								AllowedMethods: []*string{scw.StringPtr("PUT"), scw.StringPtr("POST")},
								AllowedOrigins: []*string{scw.StringPtr("https://www.example.com")},
								ExposeHeaders:  []*string{scw.StringPtr("x-amz-server-side-encryption"), scw.StringPtr("ETag")},
								MaxAgeSeconds:  scw.Int64Ptr(3000),
							},
						},
					),
				),
			},
		},
	})
}

func TestAccScalewayObjectBucket_Cors_Delete(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	bucketName := "test-acc-scaleway-object-bucket-cors-delete"
	resourceName := "scaleway_object_bucket.bucket"

	deleteBucketCors := func(tt *TestTools, n string) resource.TestCheckFunc {
		return func(s *terraform.State) error {
			rs, ok := s.RootModule().Resources[n]
			if !ok {
				return fmt.Errorf("not found: %s", n)
			}

			conn, err := newS3ClientFromMeta(tt.Meta)
			if err != nil {
				return err
			}
			_, err = conn.DeleteBucketCorsWithContext(tt.ctx, &s3.DeleteBucketCorsInput{
				Bucket: scw.StringPtr(rs.Primary.Attributes["name"]),
			})
			if err != nil && !isS3Err(err, "NoSuchCORSConfiguration", "") {
				return err
			}
			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayObjectBucketDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "bucket" {
						name = %[1]q
						cors_rule {
							allowed_headers = ["*"]
							allowed_methods = ["PUT", "POST"]
							allowed_origins = ["https://www.example.com"]
							expose_headers  = ["x-amz-server-side-encryption", "ETag"]
							max_age_seconds = 3000
						}
					}`, bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayObjectBucketExists(tt, resourceName),
					deleteBucketCors(tt, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccScalewayObjectBucket_Cors_EmptyOrigin(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	bucketName := "test-acc-scaleway-object-bucket-cors-empty-origin"
	const resourceName = "scaleway_object_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayObjectBucketDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "bucket" {
						name = %[1]q
						cors_rule {
							allowed_headers = ["*"]
							allowed_methods = ["PUT", "POST"]
							allowed_origins = [""]
							expose_headers  = ["x-amz-server-side-encryption", "ETag"]
							max_age_seconds = 3000
						}
					}`, bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayObjectBucketExists(tt, resourceName),
					testAccCheckScalewayObjectBucketCors(tt,
						resourceName,
						[]*s3.CORSRule{
							{
								AllowedHeaders: []*string{scw.StringPtr("*")},
								AllowedMethods: []*string{scw.StringPtr("PUT"), scw.StringPtr("POST")},
								AllowedOrigins: []*string{scw.StringPtr("")},
								ExposeHeaders:  []*string{scw.StringPtr("x-amz-server-side-encryption"), scw.StringPtr("ETag")},
								MaxAgeSeconds:  scw.Int64Ptr(3000),
							},
						},
					),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
		},
	})
}

func testAccCheckScalewayObjectBucketCors(tt *TestTools, n string, corsRules []*s3.CORSRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		bucketName := rs.Primary.Attributes["name"]
		s3Client, err := newS3ClientFromMeta(tt.Meta)
		if err != nil {
			return err
		}

		_, err = s3Client.HeadBucketWithContext(tt.ctx, &s3.HeadBucketInput{
			Bucket: scw.StringPtr(bucketName),
		})
		if err != nil {
			return err
		}

		out, err := s3Client.GetBucketCors(&s3.GetBucketCorsInput{
			Bucket: scw.StringPtr(bucketName),
		})
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() != "NoSuchCORSConfiguration" {
				return fmt.Errorf("GetBucketCors error: %v", err)
			}
		}

		if out == nil {
			return fmt.Errorf("CORS Rules nil")
		}

		if !reflect.DeepEqual(out.CORSRules, corsRules) {
			return fmt.Errorf("bad error cors rule, expected: %v, got %v", corsRules, out.CORSRules)
		}

		return nil
	}
}

func testAccCheckScalewayObjectBucketExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs := state.RootModule().Resources[n]
		bucketName := rs.Primary.Attributes["name"]

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

		_, err = s3Client.HeadBucket(&s3.HeadBucketInput{
			Bucket: scw.StringPtr(bucketName),
		})

		if err != nil {
			if isS3Err(err, s3.ErrCodeNoSuchBucket, "") {
				return fmt.Errorf("s3 bucket not found")
			}
			return err
		}
		return nil
	}
}
