package scaleway

import (
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("scaleway_storage_object_bucket", &resource.Sweeper{
		Name: "scaleway_storage_object_bucket",
		F:    testSweepStorageObjectBucket,
	})
}

func testSweepStorageObjectBucket(region string) error {
	s3Client, err := sharedS3ClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	log.Printf("[DEBUG] Destroying the buckets in (%s)", region)

	buckets, err := s3Client.ListBuckets()
	if err != nil {
		return fmt.Errorf("Error listing buckets in Sweeper: %s", err)
	}
	for _, bucket := range buckets {
		if strings.HasPrefix(bucket.Name, "terraform-test") {
			err := s3Client.RemoveBucket(bucket.Name)
			if err != nil {
				return fmt.Errorf("Error deleting bucket in Sweeper: %s", err)
			}
		}
	}
	return nil

}

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

		bucketExists, err := s3Client.BucketExists(rs.Primary.ID)
		if err != nil {
			return err
		}
		if bucketExists {
			return fmt.Errorf("Bucket still exists")
		}
	}

	return nil
}

var testBucketName = fmt.Sprintf("terraform-test-%d", time.Now().Unix())

var testAccCheckScalewayStorageObjectBucket = fmt.Sprintf(`
resource "scaleway_storage_object_bucket" "base" {
  name = "%s"
}
`, testBucketName)
