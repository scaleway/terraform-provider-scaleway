package scaleway

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
	awspolicy "github.com/hashicorp/awspolicyequivalence"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccScalewayBucketPolicy_Basic(t *testing.T) {
	buckedName := sdkacctest.RandomWithPrefix("tf-test-bucket")

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

	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        ErrorCheck(t, EndpointsID),
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayObjectBucketDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "bucket" {
						name = %[1]q

						tags = {
							TestName = "TestAccSCWBucketPolicy_basic"
						}
					}

					resource "scaleway_object_bucket_policy" "bucket" {
						bucket = scaleway_object_bucket.bucket.name
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
											"%[1]s",
											"%[1]s/*",
										]
										Sid = "GrantToEveryone"
									},
								]
								Version = "2012-10-17"
							}
						)
					}
					`, buckedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayObjectBucketExists(tt, "scaleway_object_bucket.bucket"),
					testAccCheckBucketHasPolicy(tt, "scaleway_object_bucket.bucket", expectedPolicyText),
				),
				ExpectNonEmptyPlan: !*UpdateCassettes,
			},
			{
				ResourceName:      "scaleway_object_bucket_policy.bucket",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckBucketHasPolicy(tt *TestTools, n string, expectedPolicyText string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		s3Client, err := newS3ClientFromMeta(tt.Meta)
		if err != nil {
			return err
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		bucketName := rs.Primary.Attributes["name"]
		policy, err := s3Client.GetBucketPolicy(&s3.GetBucketPolicyInput{
			Bucket: expandStringPtr(bucketName),
		})
		if err != nil {
			return fmt.Errorf("GetBucketPolicy error: %v", err)
		}

		actualPolicyText := *policy.Policy
		actualPolicyText, err = removePolicyStatementResources(actualPolicyText)
		if err != nil {
			return err
		}

		equivalent, err := awspolicy.PoliciesAreEquivalent(actualPolicyText, expectedPolicyText)
		if err != nil {
			return fmt.Errorf("error testing policy equivalence: %s", err)
		}
		if !equivalent {
			return fmt.Errorf("non equivalent policy error:\n\nexpected: %s\n\n     got: %s",
				expectedPolicyText, actualPolicyText)
		}

		return nil
	}
}

// remove the following:
//
//	policy["Statement"][i]["Resource"]
func removePolicyStatementResources(policy string) (string, error) {
	actualPolicyJson := make(map[string]interface{})
	json.Unmarshal([]byte(policy), &actualPolicyJson)

	if statement, ok := actualPolicyJson["Statement"].([]interface{}); ok && len(statement) > 0 {
		for _, rule := range statement {
			if rule, ok := rule.(map[string]interface{}); ok {
				delete(rule, "Resource")
			}
		}
	}

	actualPolicyTextBytes, err := json.Marshal(actualPolicyJson)
	if err != nil {
		return "", fmt.Errorf("json.Marshal error: %v", err)
	}

	return string(actualPolicyTextBytes), nil
}
