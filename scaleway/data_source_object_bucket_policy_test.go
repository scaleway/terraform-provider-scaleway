package scaleway

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayDataSourceObjectBucketPolicy_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	bucketName := sdkacctest.RandomWithPrefix("test-acc-scw-obp-data-basic")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayObjectBucketDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "main" {
						name = "%[1]s"
					}

					resource "scaleway_object_bucket_policy" "main" {
						bucket = scaleway_object_bucket.main.name
						policy = jsonencode(
							{
								Id = "MyPolicy"
								Statement = [
									{
										Action = [
											"s3:ListBucket",
											"s3:GetObject",
										]
										Effect = "Allow"
										Principal = {
											SCW = "*"
										}
										Resource  = [
											"${scaleway_object_bucket.main.name}",
											"${scaleway_object_bucket.main.name}/*",
										]
										Sid = "GrantToEveryone"
									},
								]
								Version = "2012-10-17"
							}
						)
					}

					data "scaleway_object_bucket_policy" "selected" {
						bucket = scaleway_object_bucket_policy.main.bucket
					}
				`, bucketName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaleway_object_bucket_policy.selected", "bucket", bucketName),
					resource.TestCheckResourceAttrSet("data.scaleway_object_bucket_policy.selected", "policy"),
					resource.TestCheckResourceAttrPair("data.scaleway_object_bucket_policy.selected", "policy", "scaleway_object_bucket_policy.main", "policy"),
				),
				ExpectNonEmptyPlan: !*UpdateCassettes,
			},
		},
	})
}
