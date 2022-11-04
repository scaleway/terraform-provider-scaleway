package scaleway

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
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
	testBucketACL := "private"
	testBucketUpdatedACL := "public-read"
	bucketBasic := sdkacctest.RandomWithPrefix("test-acc-scaleway-object-bucket-basic")
	bucketAms := sdkacctest.RandomWithPrefix("test-acc-scaleway-object-bucket-ams")
	bucketPar := sdkacctest.RandomWithPrefix("test-acc-scaleway-object-bucket-par")
	bucketLifecycle := sdkacctest.RandomWithPrefix("test-acc-scaleway-object-bucket-lifecycle")
	resourceNameLifecycle := "scaleway_object_bucket.par-bucket-lifecycle"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayObjectBucketDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "base-01" {
						name = "%s"
						tags = {
							foo = "bar"
						}
					}

					resource "scaleway_object_bucket" "ams-bucket-01" {
						name = "%s"
						region = "nl-ams"
						tags = {
							foo = "bar"
							baz = "qux"
						}
					}

					resource "scaleway_object_bucket" "par-bucket-01" {
						name = "%s"
						region = "fr-par"
					}
				`, bucketBasic, bucketAms, bucketPar),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_object_bucket.base-01", "name", bucketBasic),
					resource.TestCheckResourceAttr("scaleway_object_bucket.base-01", "acl", testBucketACL),
					resource.TestCheckResourceAttr("scaleway_object_bucket.base-01", "tags.%", "1"),
					resource.TestCheckResourceAttr("scaleway_object_bucket.base-01", "tags.foo", "bar"),
					resource.TestCheckResourceAttr("scaleway_object_bucket.base-01", "endpoint", fmt.Sprintf("https://%s.s3.%s.scw.cloud", bucketBasic, "fr-par")),

					resource.TestCheckResourceAttr("scaleway_object_bucket.ams-bucket-01", "name", bucketAms),
					resource.TestCheckResourceAttr("scaleway_object_bucket.ams-bucket-01", "tags.%", "2"),
					resource.TestCheckResourceAttr("scaleway_object_bucket.ams-bucket-01", "tags.foo", "bar"),
					resource.TestCheckResourceAttr("scaleway_object_bucket.ams-bucket-01", "tags.baz", "qux"),
					resource.TestCheckResourceAttr("scaleway_object_bucket.ams-bucket-01", "endpoint", fmt.Sprintf("https://%s.s3.%s.scw.cloud", bucketAms, "nl-ams")),

					resource.TestCheckResourceAttr("scaleway_object_bucket.par-bucket-01", "name", bucketPar),
					resource.TestCheckResourceAttr("scaleway_object_bucket.par-bucket-01", "endpoint", fmt.Sprintf("https://%s.s3.%s.scw.cloud", bucketPar, "fr-par")),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "base-01" {
						name = "%s"
						acl = "%s"
					}

					resource "scaleway_object_bucket" "ams-bucket-01" {
						name = "%s"
						region = "nl-ams"
						tags = {
							foo = "bar"
						}
					}
				`, bucketBasic, testBucketUpdatedACL, bucketAms),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_object_bucket.base-01", "name", bucketBasic),
					resource.TestCheckResourceAttr("scaleway_object_bucket.base-01", "acl", testBucketUpdatedACL),
					resource.TestCheckResourceAttr("scaleway_object_bucket.base-01", "tags.%", "0"),

					resource.TestCheckResourceAttr("scaleway_object_bucket.ams-bucket-01", "name", bucketAms),
					resource.TestCheckResourceAttr("scaleway_object_bucket.ams-bucket-01", "tags.%", "1"),
					resource.TestCheckResourceAttr("scaleway_object_bucket.ams-bucket-01", "tags.foo", "bar"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "par-bucket-lifecycle"{
						name = "%s"
						region = "fr-par"
						acl = "private"

						lifecycle_rule {
							id      = "id1"
							prefix  = "path1/"
							enabled = true

							expiration {
							  days = 365
							}

							transition {
							  days          = 30
							  storage_class = "STANDARD"
							}
						}
					}
				`, bucketLifecycle),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(tt, resourceNameLifecycle),
					resource.TestCheckResourceAttr("scaleway_object_bucket.par-bucket-lifecycle", "name", bucketLifecycle),
					resource.TestCheckResourceAttr(resourceNameLifecycle, "lifecycle_rule.0.id", "id1"),
					resource.TestCheckResourceAttr(resourceNameLifecycle, "lifecycle_rule.0.prefix", "path1/"),
					resource.TestCheckResourceAttr(resourceNameLifecycle, "lifecycle_rule.0.expiration.0.days", "365"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceNameLifecycle, "lifecycle_rule.0.transition.*", map[string]string{
						"days":          "30",
						"storage_class": "STANDARD",
					}),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "par-bucket-lifecycle"{
						name = "%s"
						region = "fr-par"
						acl = "private"

						lifecycle_rule {
							id      = "id1"
							prefix  = "path1/"
							enabled = true

							expiration {
							  days = 365
							}

							transition {
							  days          = 90
							  storage_class = "ONEZONE_IA"
							}
						}
					}
				`, bucketLifecycle),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(tt, resourceNameLifecycle),
					resource.TestCheckResourceAttr("scaleway_object_bucket.par-bucket-lifecycle", "name", bucketLifecycle),
					resource.TestCheckResourceAttr(resourceNameLifecycle, "lifecycle_rule.0.id", "id1"),
					resource.TestCheckResourceAttr(resourceNameLifecycle, "lifecycle_rule.0.prefix", "path1/"),
					resource.TestCheckResourceAttr(resourceNameLifecycle, "lifecycle_rule.0.expiration.0.days", "365"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceNameLifecycle, "lifecycle_rule.0.transition.*", map[string]string{
						"days":          "90",
						"storage_class": "ONEZONE_IA",
					}),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "par-bucket-lifecycle"{
						name = "%s"
						region = "fr-par"
						acl = "private"

						lifecycle_rule {
							id      = "id1"
							prefix  = "path1/"
							enabled = true

							expiration {
							  days = 365
							}

							transition {
							  days          = 120
							  storage_class = "GLACIER"
							}
						}

						lifecycle_rule {
							id      = "id2"
							prefix  = "path2/"
							enabled = true

							expiration {
							  days = "50"
							}
						}

						lifecycle_rule {
							id      = "id3"
							prefix  = "path3/"
							enabled = true

							tags = {
							  "tagKey"    = "tagValue"
							  "terraform" = "hashicorp"
							}

							expiration {
							  days = "1"
							}
						}

						lifecycle_rule {
							id      = "id4"
							enabled = true

							tags = {
							  "tagKey"    = "tagValue"
							  "terraform" = "hashicorp"
							}

							transition {
							  days          = 0
							  storage_class = "GLACIER"
							}
						}

						lifecycle_rule {
							id      = "id5"
							enabled = true

							tags = {
							  "tagKey" = "tagValue"
							}

							transition {
							  days          = 0
							  storage_class = "GLACIER"
							}
						}
					}
				`, bucketLifecycle),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(tt, resourceNameLifecycle),
					resource.TestCheckResourceAttr("scaleway_object_bucket.par-bucket-lifecycle", "name", bucketLifecycle),
					resource.TestCheckResourceAttr(resourceNameLifecycle, "lifecycle_rule.0.id", "id1"),
					resource.TestCheckResourceAttr(resourceNameLifecycle, "lifecycle_rule.0.prefix", "path1/"),
					resource.TestCheckResourceAttr(resourceNameLifecycle, "lifecycle_rule.0.expiration.0.days", "365"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceNameLifecycle, "lifecycle_rule.0.transition.*", map[string]string{
						"days":          "120",
						"storage_class": "GLACIER",
					}),
					resource.TestCheckResourceAttr(resourceNameLifecycle, "lifecycle_rule.1.id", "id2"),
					resource.TestCheckResourceAttr(resourceNameLifecycle, "lifecycle_rule.1.prefix", "path2/"),
					resource.TestCheckResourceAttr(resourceNameLifecycle, "lifecycle_rule.1.expiration.0.days", "50"),
					resource.TestCheckResourceAttr(resourceNameLifecycle, "lifecycle_rule.2.id", "id3"),
					resource.TestCheckResourceAttr(resourceNameLifecycle, "lifecycle_rule.2.prefix", "path3/"),
					resource.TestCheckResourceAttr(resourceNameLifecycle, "lifecycle_rule.2.tags.tagKey", "tagValue"),
					resource.TestCheckResourceAttr(resourceNameLifecycle, "lifecycle_rule.2.tags.terraform", "hashicorp"),
					resource.TestCheckResourceAttr(resourceNameLifecycle, "lifecycle_rule.3.id", "id4"),
					resource.TestCheckResourceAttr(resourceNameLifecycle, "lifecycle_rule.3.tags.tagKey", "tagValue"),
					resource.TestCheckResourceAttr(resourceNameLifecycle, "lifecycle_rule.3.tags.terraform", "hashicorp"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceNameLifecycle, "lifecycle_rule.3.transition.*", map[string]string{
						"days":          "0",
						"storage_class": "GLACIER",
					}),
					resource.TestCheckResourceAttr(resourceNameLifecycle, "lifecycle_rule.4.id", "id5"),
					resource.TestCheckResourceAttr(resourceNameLifecycle, "lifecycle_rule.4.tags.tagKey", "tagValue"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceNameLifecycle, "lifecycle_rule.4.transition.*", map[string]string{
						"days":          "0",
						"storage_class": "GLACIER",
					}),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "par-bucket-lifecycle"{
						name = "%s"
						region = "fr-par"
						acl = "private"

						lifecycle_rule {
							id      = "id1"
							prefix  = "path1/"
							enabled = true
							abort_incomplete_multipart_upload_days = 30
						}
					}
				`, bucketLifecycle),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(tt, resourceNameLifecycle),
					resource.TestCheckResourceAttr("scaleway_object_bucket.par-bucket-lifecycle", "name", bucketLifecycle),
					resource.TestCheckResourceAttr(resourceNameLifecycle, "lifecycle_rule.0.id", "id1"),
					resource.TestCheckResourceAttr(resourceNameLifecycle, "lifecycle_rule.0.prefix", "path1/"),
					resource.TestCheckResourceAttr(resourceNameLifecycle, "lifecycle_rule.0.abort_incomplete_multipart_upload_days", "30"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "par-bucket-lifecycle"{
						name = "%s"
						region = "fr-par"
						acl = "private"

						lifecycle_rule {
							prefix  = "path1/"
							enabled = true
							abort_incomplete_multipart_upload_days = 30
						}
					}
				`, bucketLifecycle),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(tt, resourceNameLifecycle),
					resource.TestCheckResourceAttrSet(resourceNameLifecycle, "lifecycle_rule.0.id"),
					resource.TestCheckResourceAttr(resourceNameLifecycle, "lifecycle_rule.0.abort_incomplete_multipart_upload_days", "30"),
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

	resourceName := "scaleway_object_bucket.bucket-cors-update"
	bucketName := sdkacctest.RandomWithPrefix("test-acc-scaleway-object-bucket-cors-update")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayObjectBucketDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "bucket-cors-update" {
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
							return fmt.Errorf("error putting bucket cors: %s", err)
						}
						return nil
					},
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "bucket-cors-update" {
						name = %[1]q
						cors_rule {
							allowed_headers = ["*"]
							allowed_methods = ["PUT", "POST", "GET"]
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
								AllowedMethods: []*string{scw.StringPtr("PUT"), scw.StringPtr("POST"), scw.StringPtr("GET")},
								AllowedOrigins: []*string{scw.StringPtr("https://www.example.com")},
								ExposeHeaders:  []*string{scw.StringPtr("x-amz-server-side-encryption"), scw.StringPtr("ETag")},
								MaxAgeSeconds:  scw.Int64Ptr(3000),
							},
						},
					),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "bucket-cors-update" {
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
	ctx := context.Background()

	resourceName := "scaleway_object_bucket.bucket"
	bucketName := sdkacctest.RandomWithPrefix("test-acc-scaleway-object-bucket-cors-delete")
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
			_, err = conn.DeleteBucketCorsWithContext(ctx, &s3.DeleteBucketCorsInput{
				Bucket: scw.StringPtr(rs.Primary.Attributes["name"]),
			})
			if err != nil && !isS3Err(err, ErrCodeNoSuchCORSConfiguration, "") {
				return fmt.Errorf("error deleting bucket cors: %s", err)
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

	bucketName := sdkacctest.RandomWithPrefix("test-acc-scaleway-object-cors-empty-origin")
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
				ExpectError: regexp.MustCompile("error putting S3 CORS"),
			},
		},
	})
}

func testAccCheckScalewayObjectBucketCors(tt *TestTools, n string, corsRules []*s3.CORSRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ctx := context.Background()

		rs := s.RootModule().Resources[n]
		bucketName := rs.Primary.Attributes["name"]
		s3Client, err := newS3ClientFromMeta(tt.Meta)
		if err != nil {
			return err
		}

		_, err = s3Client.HeadBucketWithContext(ctx, &s3.HeadBucketInput{
			Bucket: scw.StringPtr(bucketName),
		})
		if err != nil {
			return fmt.Errorf("error getting bucket %q: %s", bucketName, err)
		}

		out, err := s3Client.GetBucketCors(&s3.GetBucketCorsInput{
			Bucket: scw.StringPtr(bucketName),
		})
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() != ErrCodeNoSuchCORSConfiguration {
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
		if rs == nil {
			return fmt.Errorf("resource not found")
		}
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
			return fmt.Errorf("error getting s3 bucket: %v", err)
		}
		return nil
	}
}

func TestAccScalewayObjectBucket_DestroyForce(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resourceName := "scaleway_object_bucket.bucket"
	bucketName := sdkacctest.RandomWithPrefix("test-acc-scaleway-object-bucket-force")

	addObjectToBucket := func(tt *TestTools, n string) resource.TestCheckFunc {
		return func(s *terraform.State) error {
			rs, ok := s.RootModule().Resources[n]
			if !ok {
				return fmt.Errorf("not found: %s", n)
			}

			conn, err := newS3ClientFromMeta(tt.Meta)
			if err != nil {
				return err
			}
			_, err = conn.PutObject(&s3.PutObjectInput{
				Bucket: scw.StringPtr(rs.Primary.Attributes["name"]),
				Key:    scw.StringPtr("test-file"),
			})
			if err != nil {
				return fmt.Errorf("failed to put object in test bucket: %s", err)
			}
			_, err = conn.PutObject(&s3.PutObjectInput{
				Bucket: scw.StringPtr(rs.Primary.Attributes["name"]),
				Key:    scw.StringPtr("folder/test-file-in-folder"),
			})
			if err != nil {
				return fmt.Errorf("failed to put object in test bucket sub folder: %s", err)
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
						force_destroy = true
						versioning {
							enabled = true
						}
					}`, bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayObjectBucketExists(tt, resourceName),
					addObjectToBucket(tt, resourceName),
				),
			},
		},
	})
}

func testAccCheckBucketLifecycleConfigurationExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		s3Client, err := newS3ClientFromMeta(tt.Meta)
		if err != nil {
			return err
		}

		bucketRegionalID := expandRegionalID(rs.Primary.ID)

		input := &s3.GetBucketLifecycleConfigurationInput{
			Bucket: expandStringPtr(bucketRegionalID.ID),
		}

		output, err := retryOnAWSCode(context.Background(), ErrCodeNoSuchLifecycleConfiguration, func() (interface{}, error) {
			return s3Client.GetBucketLifecycleConfiguration(input) //nolint:wrapcheck
		})
		if err != nil {
			return err
		}

		if config, ok := output.(*s3.GetBucketLifecycleConfigurationOutput); !ok || config == nil {
			return fmt.Errorf("object Storage Bucket Replication Configuration for bucket (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}
