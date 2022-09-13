package scaleway

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayObjectBucketACL_Basic(t *testing.T) {
	if !*UpdateCassettes {
		t.Skip("Skipping ObjectStorage test as this kind of resource can't be deleted before 24h")
	}
	tt := NewTestTools(t)
	defer tt.Cleanup()
	testBucketName := fmt.Sprintf("terraform-test-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayObjectBucketDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "main" {
						name = "%s"
						region = "nl-ams"
					}
				
					resource "scaleway_object_bucket_acl" "main" {
						bucket = scaleway_object_bucket.main.name
						acl = "private"
						region = "nl-ams"
					}
					`, testBucketName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_object_bucket_acl.main", "bucket", testBucketName),
					resource.TestCheckResourceAttr("scaleway_object_bucket_acl.main", "acl", "private"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "main" {
						name = "%s"
						region = "nl-ams"
					}
				
					resource "scaleway_object_bucket_acl" "main" {
						bucket = scaleway_object_bucket.main.name
						acl = "public-read"
						region = "nl-ams"
					}
					`, testBucketName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_object_bucket_acl.main", "bucket", testBucketName),
					resource.TestCheckResourceAttr("scaleway_object_bucket_acl.main", "acl", "public-read"),
				),
			},
		},
	})
}
