package scaleway

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("scaleway_storage_object_bucket", &resource.Sweeper{
		Name: "scaleway_storage_object_bucket",
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
var testAccCheckScalewayStorageObjectBucket = fmt.Sprintf(`
resource "scaleway_storage_object_bucket" "base" {
  name = "%s"
  region = "nl-ams"
}

resource "scaleway_storage_object_bucket" "use-region-ams" {
  name = "%s"
  region = "nl-ams"
}

resource "scaleway_storage_object_bucket" "use-region-par" {
  name = "%s"
  region = "nl-ams"
}
`, testBucketName, testBucketNameAms, testBucketNamePar)

var testAccCheckScalewayStorageObjectBucketUpdate = fmt.Sprintf(`
resource "scaleway_storage_object_bucket" "base" {
  name = "%s"
  acl = "%s"
  region = "nl-ams"
}
`, testBucketName, testBucketUpdatedACL)

func TestAccScalewayStorageObjectBucket(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayStorageObjectBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayStorageObjectBucket,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_storage_object_bucket.base", "name", testBucketName),
					resource.TestCheckResourceAttr("scaleway_storage_object_bucket.base", "acl", testBucketACL),
					resource.TestCheckResourceAttr("scaleway_storage_object_bucket.use-region-ams", "name", testBucketNameAms),
					resource.TestCheckResourceAttr("scaleway_storage_object_bucket.use-region-par", "name", testBucketNamePar),
				),
			},
			{
				Config: testAccCheckScalewayStorageObjectBucketUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_storage_object_bucket.base", "name", testBucketName),
					resource.TestCheckResourceAttr("scaleway_storage_object_bucket.base", "acl", testBucketUpdatedACL),
				),
			},
		},
	})
}

func testAccCheckScalewayStorageObjectBucketDestroy(s *terraform.State) error {
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
