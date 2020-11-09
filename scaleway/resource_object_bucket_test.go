package scaleway

import (
	"fmt"
	"strings"
	"testing"
	"time"

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

	testBucketName := fmt.Sprintf("terraform-test-%d", time.Now().Unix())
	testBucketNameAms := testBucketName + "ams"
	testBucketNamePar := testBucketName + "par"
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
