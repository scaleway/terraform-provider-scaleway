package scaleway

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("scaleway_bucket", &resource.Sweeper{
		Name: "scaleway_bucket",
		F:    testSweepBucket,
	})
}

func testSweepBucket(region string) error {
	scaleway, err := sharedDeprecatedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	log.Printf("[DEBUG] Destroying the buckets in (%s)", region)

	containers, err := scaleway.GetContainers()
	if err != nil {
		return fmt.Errorf("Error describing buckets in Sweeper: %s", err)
	}

	for _, c := range containers {
		if err := scaleway.DeleteBucket(c.Name); err != nil {
			return fmt.Errorf("Error deleting bucket in Sweeper: %s", err)
		}
	}

	return nil
}

func TestAccScalewayBucket(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayBucket,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_bucket.base", "name", "terraform-test"),
				),
			},
		},
	})
}

func testAccCheckScalewayBucketDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*Meta).deprecatedClient

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scaleway" {
			continue
		}

		_, err := client.ListObjects(rs.Primary.ID)

		if err == nil {
			return fmt.Errorf("Bucket still exists")
		}
	}

	return nil
}

var testAccCheckScalewayBucket = `
resource "scaleway_bucket" "base" {
  name = "terraform-test"
}
`
