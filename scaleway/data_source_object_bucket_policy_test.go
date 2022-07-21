package scaleway

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
	awspolicy "github.com/hashicorp/awspolicyequivalence"
<<<<<<< HEAD
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
=======
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
>>>>>>> e0eb24f7 (feat(object): add support for bucket policy)
)

const ResourcePrefix = "tf-acc-test"

func TestAccDataSourceBucketPolicy_basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	var conf s3.GetBucketPolicyOutput
<<<<<<< HEAD
	rName := "test-acc-data-source-bucket-policy-basic"
	dataSourceName := "data.scaleway_object_bucket_policy.test"
	resourceName := "scaleway_object_bucket_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "scaleway_object_bucket" "main" {
					name = %[1]q
				}

				resource "scaleway_object_bucket_policy" "test" {
					bucket = scaleway_object_bucket.main.id
					policy = data.scaleway_object_policy_document.test.json
				}

				data "scaleway_object_policy_document" "test" {
					statement {
						effect = "Allow"

						actions = [
							"s3:GetObject",
							"s3:ListBucket",
						]

						resources = [
							scaleway_object_bucket.main.arn,
							"${scaleway_object_bucket.main.arn}/*",
						]
					}
				}
				
				data "scaleway_object_bucket_policy" "test" {
					bucket = scaleway_object_bucket.main.id
					depends_on = [scaleway_object_bucket.main]
				}`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketPolicyExists(tt, resourceName, &conf),
=======
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_s3_bucket_policy.test"
	resourceName := "aws_s3_bucket_policy.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceBucketPolicyConfigBasicConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketPolicyExists(resourceName, &conf),
>>>>>>> e0eb24f7 (feat(object): add support for bucket policy)
					testAccCheckBucketPolicyMatch(dataSourceName, "policy", resourceName, "policy"),
				),
			},
		},
	})
}

func testAccCheckBucketPolicyMatch(resource1, attr1, resource2, attr2 string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resource1]
		if !ok {
			return fmt.Errorf("not found: %s", resource1)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}
		policy1, ok := rs.Primary.Attributes[attr1]
		if !ok {
			return fmt.Errorf("attribute %q not found for %q", attr1, resource1)
		}

		rs, ok = s.RootModule().Resources[resource2]
		if !ok {
			return fmt.Errorf("not found: %s", resource2)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("mo ID is set")
		}
		policy2, ok := rs.Primary.Attributes[attr2]
		if !ok {
			return fmt.Errorf("attribute %q not found for %q", attr2, resource2)
		}

		areEquivalent, err := awspolicy.PoliciesAreEquivalent(policy1, policy2)
		if err != nil {
			return fmt.Errorf("comparing IAM Policies failed: %s", err)
		}

		if !areEquivalent {
			return fmt.Errorf("S3 bucket policies differ.\npolicy1: %s\npolicy2: %s", policy1, policy2)
		}

		return nil
	}
}

<<<<<<< HEAD
func testAccCheckBucketPolicyExists(tt *TestTools, n string, ci *s3.GetBucketPolicyOutput) resource.TestCheckFunc {
=======
func testAccCheckBucketPolicyExists(n string, ci *s3.GetBucketPolicyOutput) resource.TestCheckFunc {
>>>>>>> e0eb24f7 (feat(object): add support for bucket policy)
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no S3 Bucket Policy ID is set")
		}

<<<<<<< HEAD
		s3Client, err := newS3ClientFromMeta(tt.Meta)
		if err != nil {
			return err
		}

		output, err := FindBucketPolicy(context.Background(), s3Client, rs.Primary.ID)
=======
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn

		output, err := tfs3.FindBucketPolicy(context.Background(), conn, rs.Primary.ID)
>>>>>>> e0eb24f7 (feat(object): add support for bucket policy)
		if err != nil {
			return err
		}

		*ci = *output

		return nil
	}
}
<<<<<<< HEAD
=======

func testAccDataSourceBucketPolicyBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
  tags = {
    Name = %[1]q
  }
}
resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.id
  policy = data.aws_iam_policy_document.test.json
}
data "aws_iam_policy_document" "test" {
  statement {
    effect = "Allow"
    actions = [
      "s3:GetObject",
      "s3:ListBucket",
    ]
    resources = [
      aws_s3_bucket.test.arn,
      "${aws_s3_bucket.test.arn}/*",
    ]
    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}
`, rName)
}

func testAccDataSourceBucketPolicyConfigBasicConfig(rName string) string {
	return acctest.ConfigCompose(testAccDataSourceBucketPolicyBaseConfig(rName), `
data "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.id
  depends_on = [aws_s3_bucket_policy.test]
}
`)
}
>>>>>>> e0eb24f7 (feat(object): add support for bucket policy)
