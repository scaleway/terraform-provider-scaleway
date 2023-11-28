package scaleway

import (
	"fmt"
	"testing"

	awspolicy "github.com/hashicorp/awspolicyequivalence"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccScalewayDataSourceObjectBucketPolicy_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	bucketName := sdkacctest.RandomWithPrefix("test-acc-scw-obp-data-basic")

	expectedPolicyText := `{
	"Version":"2012-10-17",
	"Id":"MyPolicy",
	"Statement": [
		{
			"Sid":"GrantToEveryone",
			"Effect":"Allow",
			"Principal":{
				"SCW":"*"
			},
			"Action":[
				"s3:ListBucket",
				"s3:GetObject"
			]
		}
   ]
}`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayObjectBucketDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "main" {
						name = "%[1]s"
						region = "%[2]s"
					}

					resource "scaleway_object_bucket_policy" "main" {
						bucket = scaleway_object_bucket.main.id
						region = "%[2]s"
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
				`, bucketName, objectTestsMainRegion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayObjectBucketExistsForceRegion(tt, "scaleway_object_bucket.main", true),
					resource.TestCheckResourceAttr("data.scaleway_object_bucket_policy.selected", "bucket", objectTestsMainRegion+"/"+bucketName),
					resource.TestCheckResourceAttrSet("data.scaleway_object_bucket_policy.selected", "policy"),
					testAccCheckDataSourcePolicyIsEquivalent("data.scaleway_object_bucket_policy.selected", expectedPolicyText),
				),
				ExpectNonEmptyPlan: !*UpdateCassettes,
			},
		},
	})
}

func testAccCheckDataSourcePolicyIsEquivalent(n, expectedPolicyText string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ds, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		dataSourcePolicy := ds.Primary.Attributes["policy"]

		dataSourcePolicyToCompare, err := removePolicyStatementResources(dataSourcePolicy)
		if err != nil {
			return err
		}

		equivalent, err := awspolicy.PoliciesAreEquivalent(expectedPolicyText, dataSourcePolicyToCompare)
		if err != nil {
			return fmt.Errorf("error testing policy equivalence: %s", err)
		}
		if !equivalent {
			return fmt.Errorf("non equivalent policy error:\n\nexpected: %s\n\n     got: %s",
				expectedPolicyText, dataSourcePolicyToCompare)
		}

		return nil
	}
}
