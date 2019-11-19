package scaleway

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func init() {
	resource.AddTestSweepers("scaleway_object_bucket", &resource.Sweeper{
		Name: "scaleway_object_bucket",
		F:    testSweepStorageObjectBucket,
	})
}

// Test data
var (
	testBucketName       = fmt.Sprintf("terraform-test-%d", time.Now().Unix())
	testBucketNameAms    = testBucketName + "ams"
	testBucketNamePar    = testBucketName + "par"
	testBucketACL        = "private"
	testBucketUpdatedACL = "public-read"
)

// Test configs
var testAccCheckScalewayObjectBucket = fmt.Sprintf(`
	resource "scaleway_object_bucket" "base" {
		name = "%s"
	}

	resource "scaleway_object_bucket" "ams-bucket" {
		name = "%s"
		region = "nl-ams"
	}

	resource "scaleway_object_bucket" "par-bucket" {
		name = "%s"
		region = "fr-par"
	}
`, testBucketName, testBucketNameAms, testBucketNamePar)

var testAccCheckScalewayObjectBucketUpdate = fmt.Sprintf(`
	resource "scaleway_object_bucket" "base" {
		name = "%s"
		acl = "%s"
	}
`, testBucketName, testBucketUpdatedACL)

func TestAccScalewayObjectBucket(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayObjectBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayObjectBucket,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_object_bucket.base", "name", testBucketName),
					resource.TestCheckResourceAttr("scaleway_object_bucket.base", "acl", testBucketACL),
					resource.TestCheckResourceAttr("scaleway_object_bucket.ams-bucket", "name", testBucketNameAms),
					resource.TestCheckResourceAttr("scaleway_object_bucket.par-bucket", "name", testBucketNamePar),
				),
			},
			{
				Config: testAccCheckScalewayObjectBucketUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_object_bucket.base", "name", testBucketName),
					resource.TestCheckResourceAttr("scaleway_object_bucket.base", "acl", testBucketUpdatedACL),
				),
			},
		},
	})
}

func testAccCheckScalewayObjectBucketDestroy(s *terraform.State) error {
	s3Client := testAccProvider.Meta().(*Meta).s3Client

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scaleway" {
			continue
		}

		bucketName := rs.Primary.ID

		_, err := s3Client.ListObjects(&s3.ListObjectsInput{
			Bucket: &bucketName,
		})
		if err != nil {
			if serr, ok := err.(awserr.Error); ok && serr.Code() == s3.ErrCodeNoSuchBucket {
				// bucket doesn't exist
				continue
			}
			return fmt.Errorf("couldn't get bucket to verify if it stil exists: %s", err)
		}

		return fmt.Errorf("bucket should be deleted")
	}

	return nil
}

func testSweepStorageObjectBucket(region string) error {
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
				return fmt.Errorf("Error deleting bucket in Sweeper: %s", err)
			}
		}

	}

	return nil

}
