package scaleway

import (
<<<<<<< HEAD
=======
<<<<<<< HEAD
	"encoding/json"
>>>>>>> 65b6efaa (feat(object): add support for bucket policy)
	"fmt"
	"testing"
	awspolicy "github.com/hashicorp/awspolicyequivalence"

	"github.com/aws/aws-sdk-go/aws"
<<<<<<< HEAD
=======
	"github.com/aws/aws-sdk-go/aws/awserr"
=======
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
>>>>>>> e0eb24f7 (feat(object): add support for bucket policy)
>>>>>>> 65b6efaa (feat(object): add support for bucket policy)
	"github.com/aws/aws-sdk-go/service/s3"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
<<<<<<< HEAD
)

func TestAccS3BucketPolicy_basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
=======
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
)

func TestAccS3BucketPolicy_basic(t *testing.T) {
	name := fmt.Sprintf("tf-test-bucket-%d", sdkacctest.RandInt())
	partition := acctest.Partition()
>>>>>>> e0eb24f7 (feat(object): add support for bucket policy)

	bucketName := fmt.Sprintf("tf-test-bucket-%d", sdkacctest.RandInt())

	expectedPolicyText := fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
<<<<<<< HEAD
      "Principal": "*",
      "Action": "s3:*",
      "Resource": [
        "%[1]s/*",
        "%[1]s"
      ]
    }
  ]
}`, bucketName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayObjectBucketDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketPolicyConfig(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayObjectBucketExists(tt, "scaleway_object_bucket.bucket"),
					testAccCheckBucketHasPolicy(tt, "scaleway_object_bucket.bucket", expectedPolicyText),
				),
			},
			{
				ResourceName:      "scaleway_object_bucket_policy.bucket",
=======
      "Principal": {
        "AWS": "*"
      },
      "Action": "s3:*",
      "Resource": [
        "arn:%s:s3:::%s/*",
        "arn:%s:s3:::%s"
      ]
    }
  ]
}`, partition, name, partition, name)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketPolicyConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists("aws_s3_bucket.bucket"),
					testAccCheckBucketHasPolicy("aws_s3_bucket.bucket", expectedPolicyText),
				),
			},
			{
				ResourceName:      "aws_s3_bucket_policy.bucket",
>>>>>>> e0eb24f7 (feat(object): add support for bucket policy)
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccS3BucketPolicy_disappears(t *testing.T) {
<<<<<<< HEAD
	tt := NewTestTools(t)
	defer tt.Cleanup()

=======
<<<<<<< HEAD
>>>>>>> 65b6efaa (feat(object): add support for bucket policy)
	name := "test-acc-s3-bucket-policy-disappears"
	bucketResourceName := "scaleway_object_bucket.bucket"
=======
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	partition := acctest.Partition()
	bucketResourceName := "aws_s3_bucket.bucket"
	resourceName := "aws_s3_bucket_policy.bucket"
>>>>>>> e0eb24f7 (feat(object): add support for bucket policy)

	expectedPolicyText := fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
<<<<<<< HEAD
        "SCW": "*"
      },
      "Action": "s3:*",
      "Resource": [
        "%[1]s/*",
        "%[1]s"
      ]
    }
  ]
<<<<<<< HEAD
}`, name, name)
=======
}`, name)
	tt := NewTestTools(t)
	defer tt.Cleanup()
>>>>>>> c2713163 (Fix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayObjectBucketDestroy(tt),
=======
        "AWS": "*"
      },
      "Action": "s3:*",
      "Resource": [
        "arn:%s:s3:::%s/*",
        "arn:%s:s3:::%s"
      ]
    }
  ]
}`, partition, name, partition, name)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
>>>>>>> e0eb24f7 (feat(object): add support for bucket policy)
		Steps: []resource.TestStep{
			{
				Config: testAccBucketPolicyConfig(name),
				Check: resource.ComposeTestCheckFunc(
<<<<<<< HEAD
					testAccCheckScalewayObjectBucketExists(tt, bucketResourceName),
					testAccCheckBucketHasPolicy(tt, bucketResourceName, expectedPolicyText),
=======
					testAccCheckBucketExists(bucketResourceName),
					testAccCheckBucketHasPolicy(bucketResourceName, expectedPolicyText),
					acctest.CheckResourceDisappears(acctest.Provider, tfs3.ResourceBucketPolicy(), resourceName),
>>>>>>> e0eb24f7 (feat(object): add support for bucket policy)
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3BucketPolicy_disappears_bucket(t *testing.T) {
<<<<<<< HEAD
	tt := NewTestTools(t)
	defer tt.Cleanup()

=======
<<<<<<< HEAD
>>>>>>> 65b6efaa (feat(object): add support for bucket policy)
	name := "test-acc-s3-bucket-policy-disappears-bucket"
	bucketResourceName := "scaleway_object_bucket.bucket"
=======
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	partition := acctest.Partition()
	bucketResourceName := "aws_s3_bucket.bucket"
>>>>>>> e0eb24f7 (feat(object): add support for bucket policy)

	expectedPolicyText := fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
<<<<<<< HEAD
        "SCW": "*"
      },
      "Action": "s3:*",
      "Resource": [
        "%[1]s/*",
        "%[1]s"
      ]
    }
  ]
<<<<<<< HEAD
}`, name, name)
=======
}`, name)
	tt := NewTestTools(t)
	defer tt.Cleanup()
>>>>>>> c2713163 (Fix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayObjectBucketDestroy(tt),
=======
        "AWS": "*"
      },
      "Action": "s3:*",
      "Resource": [
        "arn:%s:s3:::%s/*",
        "arn:%s:s3:::%s"
      ]
    }
  ]
}`, partition, name, partition, name)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
>>>>>>> e0eb24f7 (feat(object): add support for bucket policy)
		Steps: []resource.TestStep{
			{
				Config: testAccBucketPolicyConfig(name),
				Check: resource.ComposeTestCheckFunc(
<<<<<<< HEAD
					testAccCheckScalewayObjectBucketExists(tt, bucketResourceName),
					testAccCheckBucketHasPolicy(tt, bucketResourceName, expectedPolicyText),
=======
					testAccCheckBucketExists(bucketResourceName),
					testAccCheckBucketHasPolicy(bucketResourceName, expectedPolicyText),
					acctest.CheckResourceDisappears(acctest.Provider, tfs3.ResourceBucket(), bucketResourceName),
>>>>>>> e0eb24f7 (feat(object): add support for bucket policy)
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3BucketPolicy_policyUpdate(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	name := fmt.Sprintf("tf-test-bucket-%d", sdkacctest.RandInt())
<<<<<<< HEAD
=======
	partition := acctest.Partition()
>>>>>>> e0eb24f7 (feat(object): add support for bucket policy)

	expectedPolicyText1 := fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
<<<<<<< HEAD
        "SCW": "*"
      },
      "Action": "s3:*",
      "Resource": [
        "%[1]s/*",
        "%[1]s"
      ]
    }
  ]
}`, name)
=======
        "AWS": "*"
      },
      "Action": "s3:*",
      "Resource": [
        "arn:%[1]s:s3:::%[2]s/*",
        "arn:%[1]s:s3:::%[2]s"
      ]
    }
  ]
}`, partition, name)
>>>>>>> e0eb24f7 (feat(object): add support for bucket policy)

	expectedPolicyText2 := fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
<<<<<<< HEAD
        "SCW": "*"
=======
        "AWS": "*"
>>>>>>> e0eb24f7 (feat(object): add support for bucket policy)
      },
      "Action": [
        "s3:DeleteBucket",
        "s3:ListBucket",
        "s3:ListBucketVersions"
      ],
      "Resource": [
<<<<<<< HEAD
        "%[1]s/*",
        "%[1]s"
      ]
    }
  ]
}`, name)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: tt.ProviderFactories,
<<<<<<< HEAD
		CheckDestroy:      testAccCheckScalewayObjectBucketDestroy(tt),
=======

		CheckDestroy: testAccCheckScalewayObjectBucketDestroy(tt),
=======
        "arn:%[1]s:s3:::%[2]s/*",
        "arn:%[1]s:s3:::%[2]s"
      ]
    }
  ]
}`, partition, name)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
>>>>>>> e0eb24f7 (feat(object): add support for bucket policy)
>>>>>>> 65b6efaa (feat(object): add support for bucket policy)
		Steps: []resource.TestStep{
			{
				Config: testAccBucketPolicyConfig(name),
				Check: resource.ComposeTestCheckFunc(
<<<<<<< HEAD
					testAccCheckScalewayObjectBucketExists(tt, "scaleway_object_bucket.bucket"),
					testAccCheckBucketHasPolicy(tt, "scaleway_object_bucket.bucket", expectedPolicyText1),
=======
					testAccCheckBucketExists("aws_s3_bucket.bucket"),
					testAccCheckBucketHasPolicy("aws_s3_bucket.bucket", expectedPolicyText1),
>>>>>>> e0eb24f7 (feat(object): add support for bucket policy)
				),
			},

			{
<<<<<<< HEAD
				Config: testaccbucketpolicyconfigUpdated(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayObjectBucketExists(tt, "scaleway_object_bucket.bucket"),
					testAccCheckBucketHasPolicy(tt, "scaleway_object_bucket.bucket", expectedPolicyText2),
=======
				Config: testAccBucketPolicyConfig_updated(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists("aws_s3_bucket.bucket"),
					testAccCheckBucketHasPolicy("aws_s3_bucket.bucket", expectedPolicyText2),
>>>>>>> e0eb24f7 (feat(object): add support for bucket policy)
				),
			},

			{
<<<<<<< HEAD
				ResourceName:      "scaleway_object_bucket_policy.bucket",
=======
				ResourceName:      "aws_s3_bucket_policy.bucket",
>>>>>>> e0eb24f7 (feat(object): add support for bucket policy)
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

<<<<<<< HEAD
func TestAccS3BucketPolicy_migrate_noChange(t *testing.T) {
	rName := "test-acc-s3-bucket-policy-migrate-noChange"
	resourceName := "scaleway_object_bucket_policy.test"
	bucketResourceName := "scaleway_object_bucket.test"

	tt := NewTestTools(t)
	defer tt.Cleanup()

	rName := sdkacctest.RandomWithPrefix(ResourcePrefix)
	resourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayObjectBucketDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketPolicyIAMRoleOrderIAMPolicyDocConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayObjectBucketExists(tt, resourceName),
				),
			},
			{
				Config:   testAccBucketPolicyIAMRoleOrderIAMPolicyDocConfig(rName),
				PlanOnly: true,
			},
			{
				Config:   testAccBucketPolicyIAMRoleOrderIAMPolicyDocConfig(rName),
				PlanOnly: true,
			},
			{
				Config:   testAccBucketPolicyIAMRoleOrderIAMPolicyDocConfig(rName),
				PlanOnly: true,
				Config: testaccbucketconfigWithpolicy(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScalewayObjectBucketExists(tt, bucketResourceName),
					testAccCheckBucketPolicy(tt, bucketResourceName, testAccBucketPolicy(rName)),
				),
			},
			{
				Config: testaccbucketpolicyMigrateNochangeconfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScalewayObjectBucketExists(tt, bucketResourceName),
					testAccCheckBucketPolicy(tt, resourceName, testAccBucketPolicy(rName)),
=======
// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/11801
func TestAccS3BucketPolicy_IAMRoleOrder_policyDoc(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketPolicyIAMRoleOrderIAMPolicyDocConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
				),
			},
			{
				Config:   testAccBucketPolicyIAMRoleOrderIAMPolicyDocConfig(rName),
				PlanOnly: true,
			},
			{
				Config:   testAccBucketPolicyIAMRoleOrderIAMPolicyDocConfig(rName),
				PlanOnly: true,
			},
			{
				Config:   testAccBucketPolicyIAMRoleOrderIAMPolicyDocConfig(rName),
				PlanOnly: true,
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/13144
// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/20456
func TestAccS3BucketPolicy_IAMRoleOrder_policyDocNotPrincipal(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketPolicyIAMRoleOrderIAMPolicyDocNotPrincipalConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
				),
			},
			{
				Config: testAccBucketPolicyIAMRoleOrderIAMPolicyDocNotPrincipalConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
				),
			},
			{
				Config:   testAccBucketPolicyIAMRoleOrderIAMPolicyDocNotPrincipalConfig(rName),
				PlanOnly: true,
			},
			{
				Config:   testAccBucketPolicyIAMRoleOrderIAMPolicyDocNotPrincipalConfig(rName),
				PlanOnly: true,
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/11801
func TestAccS3BucketPolicy_IAMRoleOrder_jsonEncode(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName3 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketPolicyIAMRoleOrderJSONEncodeConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
				),
			},
			{
				Config:   testAccBucketPolicyIAMRoleOrderJSONEncodeConfig(rName),
				PlanOnly: true,
			},
			{
				Config: testAccBucketPolicyIAMRoleOrderJSONEncodeOrder2Config(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
				),
			},
			{
				Config:   testAccBucketPolicyIAMRoleOrderJSONEncodeConfig(rName2),
				PlanOnly: true,
			},
			{
				Config: testAccBucketPolicyIAMRoleOrderJSONEncodeOrder3Config(rName3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
				),
			},
			{
				Config:   testAccBucketPolicyIAMRoleOrderJSONEncodeConfig(rName3),
				PlanOnly: true,
			},
		},
	})
}

func TestAccS3BucketPolicy_migrate_noChange(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_policy.test"
	bucketResourceName := "aws_s3_bucket.test"
	partition := acctest.Partition()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_withPolicy(rName, partition),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(bucketResourceName),
					testAccCheckBucketPolicy(bucketResourceName, testAccBucketPolicy(rName, partition)),
				),
			},
			{
				Config: testAccBucketPolicy_Migrate_NoChangeConfig(rName, partition),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(bucketResourceName),
					testAccCheckBucketPolicy(resourceName, testAccBucketPolicy(rName, partition)),
>>>>>>> e0eb24f7 (feat(object): add support for bucket policy)
				),
>>>>>>> c2713163 (Fix)
			},
		},
	})
}

<<<<<<< HEAD
// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/13144
// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/20456
func TestAccS3BucketPolicy_IAMRoleOrder_policyDocNotPrincipal(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(ResourcePrefix)
	resourceName := "aws_s3_bucket.test"
=======
func TestAccS3BucketPolicy_migrate_withChange(t *testing.T) {
<<<<<<< HEAD
	rName := "test-acc-s3-bucket-policy-migrate-with-change"
	resourceName := "scaleway_object_bucket_policy.test"
	bucketResourceName := "scaleway_object_bucket.test"

>>>>>>> c2713163 (Fix)
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayObjectBucketDestroy(tt),
		Steps: []resource.TestStep{
			{
<<<<<<< HEAD
				Config: testAccBucketPolicyIAMRoleOrderIAMPolicyDocNotPrincipalConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayObjectBucketExists(tt, resourceName),
				),
			},
			{
				Config: testAccBucketPolicyIAMRoleOrderIAMPolicyDocNotPrincipalConfig(rName),
				Check: resource.ComposeTestCheckFunc(
=======
				Config: testaccbucketconfigWithpolicy(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScalewayObjectBucketExists(tt, bucketResourceName),
					testAccCheckBucketPolicy(tt, bucketResourceName, testAccBucketPolicy(rName)),
				),
			},
			{
				Config: testaccbucketpolicyMigrateWithchangeconfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
>>>>>>> c2713163 (Fix)
					testAccCheckScalewayObjectBucketExists(tt, resourceName),
<<<<<<< HEAD
=======
					testAccCheckBucketPolicy(tt, resourceName, testAccBucketPolicyUpdated(rName)),
=======
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_policy.test"
	bucketResourceName := "aws_s3_bucket.test"
	partition := acctest.Partition()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_withPolicy(rName, partition),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(bucketResourceName),
					testAccCheckBucketPolicy(bucketResourceName, testAccBucketPolicy(rName, partition)),
				),
			},
			{
				Config: testAccBucketPolicy_Migrate_WithChangeConfig(rName, partition),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					testAccCheckBucketPolicy(resourceName, testAccBucketPolicyUpdated(rName, partition)),
>>>>>>> e0eb24f7 (feat(object): add support for bucket policy)
>>>>>>> 65b6efaa (feat(object): add support for bucket policy)
				),
			},
			{
				Config:   testAccBucketPolicyIAMRoleOrderIAMPolicyDocNotPrincipalConfig(rName),
				PlanOnly: true,
			},
			{
				Config:   testAccBucketPolicyIAMRoleOrderIAMPolicyDocNotPrincipalConfig(rName),
				PlanOnly: true,
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/11801
func TestAccS3BucketPolicy_IAMRoleOrder_jsonEncode(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	rName := sdkacctest.RandomWithPrefix(ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(ResourcePrefix)
	rName3 := sdkacctest.RandomWithPrefix(ResourcePrefix)
	resourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayObjectBucketDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketPolicyIAMRoleOrderJSONEncodeConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayObjectBucketExists(tt, resourceName),
				),
			},
			{
				Config:   testAccBucketPolicyIAMRoleOrderJSONEncodeConfig(rName),
				PlanOnly: true,
			},
			{
				Config: testAccBucketPolicyIAMRoleOrderJSONEncodeOrder2Config(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayObjectBucketExists(tt, resourceName),
				),
			},
			{
				Config:   testAccBucketPolicyIAMRoleOrderJSONEncodeConfig(rName2),
				PlanOnly: true,
			},
			{
				Config: testAccBucketPolicyIAMRoleOrderJSONEncodeOrder3Config(rName3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayObjectBucketExists(tt, resourceName),
				),
			},
			{
				Config:   testAccBucketPolicyIAMRoleOrderJSONEncodeConfig(rName3),
				PlanOnly: true,
			},
		},
	})
}

<<<<<<< HEAD
func testAccCheckBucketHasPolicy(tt *TestTools, n string, expectedPolicyText string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no S3 Bucket ID is set")
		}

		s3Client, err := newS3ClientFromMeta(tt.Meta)
		if err != nil {
			return err
		}

		policy, err := s3Client.GetBucketPolicy(&s3.GetBucketPolicyInput{
=======
func testAccCheckBucketHasPolicy(n string, expectedPolicyText string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No S3 Bucket ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn

		policy, err := conn.GetBucketPolicy(&s3.GetBucketPolicyInput{
>>>>>>> e0eb24f7 (feat(object): add support for bucket policy)
			Bucket: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return fmt.Errorf("GetBucketPolicy error: %v", err)
		}

		actualPolicyText := *policy.Policy

		equivalent, err := awspolicy.PoliciesAreEquivalent(actualPolicyText, expectedPolicyText)
		if err != nil {
<<<<<<< HEAD
			return fmt.Errorf("error testing policy equivalence: %s", err)
		}
		if !equivalent {
			return fmt.Errorf("non-equivalent policy error:\n\nexpected: %s\n\n     got: %s",
=======
			return fmt.Errorf("Error testing policy equivalence: %s", err)
		}
		if !equivalent {
			return fmt.Errorf("Non-equivalent policy error:\n\nexpected: %s\n\n     got: %s\n",
>>>>>>> e0eb24f7 (feat(object): add support for bucket policy)
				expectedPolicyText, actualPolicyText)
		}

		return nil
	}
}

func testAccBucketPolicyConfig(bucketName string) string {
	return fmt.Sprintf(`
<<<<<<< HEAD
resource "scaleway_object_bucket" "bucket" {
  name = %[1]q
=======
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
>>>>>>> e0eb24f7 (feat(object): add support for bucket policy)
  tags = {
    TestName = "TestAccS3BucketPolicy_basic"
  }
}
<<<<<<< HEAD

resource "scaleway_object_bucket_policy" "bucket" {
  bucket = scaleway_object_bucket.bucket.name
  policy = data.scaleway_object_policy_document.policy.json
}

data "scaleway_object_policy_document" "policy" {
=======
resource "aws_s3_bucket_policy" "bucket" {
  bucket = aws_s3_bucket.bucket.bucket
  policy = data.aws_iam_policy_document.policy.json
}
data "aws_iam_policy_document" "policy" {
>>>>>>> e0eb24f7 (feat(object): add support for bucket policy)
  statement {
    effect = "Allow"
    actions = [
      "s3:*",
    ]
    resources = [
<<<<<<< HEAD
      scaleway_object_bucket.bucket.name,
      "${scaleway_object_bucket.bucket.name}/*",
    ]
    principals {
      type        = "SCW"
      identifiers = ["*"]
    }
  }
}
`, bucketName)
}

func testaccbucketpolicyconfigUpdated(bucketName string) string {
	return fmt.Sprintf(`
resource "scaleway_object_bucket" "bucket" {
  name = %[1]q
  tags = {
    TestName = "TestAccS3BucketPolicy_basic"
  }
}

resource "scaleway_object_bucket_policy" "bucket" {
  bucket = scaleway_object_bucket.bucket.name
  policy = data.scaleway_object_policy_document.policy.json
}

data "scaleway_object_policy_document" "policy" {
  statement {
    effect = "Allow"
    actions = [
      "s3:DeleteBucket",
      "s3:ListBucket",
      "s3:ListBucketVersions",
    ]
    resources = [
      scaleway_object_bucket.bucket.name,
      "${scaleway_object_bucket.bucket.name}/*",
=======
      aws_s3_bucket.bucket.arn,
      "${aws_s3_bucket.bucket.arn}/*",
>>>>>>> e0eb24f7 (feat(object): add support for bucket policy)
    ]
    principals {
      type        = "AWS"
      identifiers = ["*"]
    }
  }
}
`, bucketName)
}

<<<<<<< HEAD
<<<<<<< HEAD
func testAccBucketPolicyIAMRoleOrderBaseConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}
resource "aws_iam_role" "test1" {
  name = "%[1]s-sultan"
  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "s3.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}
resource "aws_iam_role" "test2" {
  name = "%[1]s-shepard"
  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "s3.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}
resource "aws_iam_role" "test3" {
  name = "%[1]s-tritonal"
  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "s3.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}
resource "aws_iam_role" "test4" {
  name = "%[1]s-artlec"
  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "s3.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}
resource "aws_iam_role" "test5" {
  name = "%[1]s-cazzette"
  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "s3.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
  tags = {
    TestName = %[1]q
  }
}
`, rName)
}

func testAccBucketPolicyIAMRoleOrderIAMPolicyDocConfig(rName string) string {
	return ConfigCompose(
		testAccBucketPolicyIAMRoleOrderBaseConfig(rName),
		fmt.Sprintf(`
data "aws_iam_policy_document" "test" {
  policy_id = %[1]q
  statement {
    actions = [
      "s3:DeleteBucket",
      "s3:ListBucket",
      "s3:ListBucketVersions",
    ]
    effect = "Allow"
    principals {
      identifiers = [
        aws_iam_role.test2.arn,
        aws_iam_role.test1.arn,
        aws_iam_role.test4.arn,
        aws_iam_role.test3.arn,
        aws_iam_role.test5.arn,
      ]
      type = "AWS"
=======
=======
>>>>>>> 65b6efaa (feat(object): add support for bucket policy)
func testaccbucketpolicyMigrateNochangeconfig(bucketName string) string {
	return fmt.Sprintf(`
resource "scaleway_object_bucket" "test" {
  name = %[1]q
}

resource "scaleway_object_bucket_acl" "test" {
  bucket = scaleway_object_bucket.test.id
  acl    = "private"
}

resource "scaleway_object_bucket_policy" "test" {
  bucket = scaleway_object_bucket.test.id
  policy = %[2]s
}
`, bucketName, strconv.Quote(testAccBucketPolicy(bucketName)))
}

func testaccbucketpolicyMigrateWithchangeconfig(bucketName string) string {
	return fmt.Sprintf(`
resource "scaleway_object_bucket" "test" {
  name = %[1]q
}

resource "scaleway_object_bucket_acl" "test" {
  bucket = scaleway_object_bucket.test.id
  acl    = "private"
}

resource "scaleway_object_bucket_policy" "test" {
  bucket = scaleway_object_bucket.test.id
  policy = %[2]s
}
`, bucketName, strconv.Quote(testAccBucketPolicyUpdated(bucketName)))
}

func testAccBucketPolicyUpdated(bucketName string) string {
=======
func testAccBucketPolicyConfig_updated(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
  tags = {
    TestName = "TestAccS3BucketPolicy_basic"
  }
}
resource "aws_s3_bucket_policy" "bucket" {
  bucket = aws_s3_bucket.bucket.bucket
  policy = data.aws_iam_policy_document.policy.json
}
data "aws_iam_policy_document" "policy" {
  statement {
    effect = "Allow"
    actions = [
      "s3:DeleteBucket",
      "s3:ListBucket",
      "s3:ListBucketVersions",
    ]
    resources = [
      aws_s3_bucket.bucket.arn,
      "${aws_s3_bucket.bucket.arn}/*",
    ]
    principals {
      type        = "AWS"
      identifiers = ["*"]
    }
  }
}
`, bucketName)
}

func testAccBucketPolicyIAMRoleOrderBaseConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}
resource "aws_iam_role" "test1" {
  name = "%[1]s-sultan"
  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "s3.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}
resource "aws_iam_role" "test2" {
  name = "%[1]s-shepard"
  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "s3.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}
resource "aws_iam_role" "test3" {
  name = "%[1]s-tritonal"
  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "s3.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}
resource "aws_iam_role" "test4" {
  name = "%[1]s-artlec"
  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "s3.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}
resource "aws_iam_role" "test5" {
  name = "%[1]s-cazzette"
  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "s3.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
  tags = {
    TestName = %[1]q
  }
}
`, rName)
}

func testAccBucketPolicyIAMRoleOrderIAMPolicyDocConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccBucketPolicyIAMRoleOrderBaseConfig(rName),
		fmt.Sprintf(`
data "aws_iam_policy_document" "test" {
  policy_id = %[1]q
  statement {
    actions = [
      "s3:DeleteBucket",
      "s3:ListBucket",
      "s3:ListBucketVersions",
    ]
    effect = "Allow"
    principals {
      identifiers = [
        aws_iam_role.test2.arn,
        aws_iam_role.test1.arn,
        aws_iam_role.test4.arn,
        aws_iam_role.test3.arn,
        aws_iam_role.test5.arn,
      ]
      type = "AWS"
    }
    resources = [
      aws_s3_bucket.test.arn,
      "${aws_s3_bucket.test.arn}/*",
    ]
  }
}
resource "aws_s3_bucket_policy" "bucket" {
  bucket = aws_s3_bucket.test.bucket
  policy = data.aws_iam_policy_document.test.json
}
`, rName))
}

func testAccBucketPolicyIAMRoleOrderJSONEncodeConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccBucketPolicyIAMRoleOrderBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket_policy" "bucket" {
  bucket = aws_s3_bucket.test.bucket
  policy = jsonencode({
    Id = %[1]q
    Statement = [{
      Action = [
        "s3:DeleteBucket",
        "s3:ListBucket",
        "s3:ListBucketVersions",
      ]
      Effect = "Allow"
      Principal = {
        AWS = [
          aws_iam_role.test2.arn,
          aws_iam_role.test1.arn,
          aws_iam_role.test4.arn,
          aws_iam_role.test3.arn,
          aws_iam_role.test5.arn,
        ]
      }
      Resource = [
        aws_s3_bucket.test.arn,
        "${aws_s3_bucket.test.arn}/*",
      ]
    }]
    Version = "2012-10-17"
  })
}
`, rName))
}

func testAccBucketPolicyIAMRoleOrderJSONEncodeOrder2Config(rName string) string {
	return acctest.ConfigCompose(
		testAccBucketPolicyIAMRoleOrderBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket_policy" "bucket" {
  bucket = aws_s3_bucket.test.bucket
  policy = jsonencode({
    Id = %[1]q
    Statement = [{
      Action = [
        "s3:DeleteBucket",
        "s3:ListBucket",
        "s3:ListBucketVersions",
      ]
      Effect = "Allow"
      Principal = {
        AWS = [
          aws_iam_role.test2.arn,
          aws_iam_role.test3.arn,
          aws_iam_role.test5.arn,
          aws_iam_role.test1.arn,
          aws_iam_role.test4.arn,
        ]
      }
      Resource = [
        aws_s3_bucket.test.arn,
        "${aws_s3_bucket.test.arn}/*",
      ]
    }]
    Version = "2012-10-17"
  })
}
`, rName))
}

func testAccBucketPolicyIAMRoleOrderJSONEncodeOrder3Config(rName string) string {
	return acctest.ConfigCompose(
		testAccBucketPolicyIAMRoleOrderBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket_policy" "bucket" {
  bucket = aws_s3_bucket.test.bucket
  policy = jsonencode({
    Id = %[1]q
    Statement = [{
      Action = [
        "s3:DeleteBucket",
        "s3:ListBucket",
        "s3:ListBucketVersions",
      ]
      Effect = "Allow"
      Principal = {
        AWS = [
          aws_iam_role.test4.arn,
          aws_iam_role.test1.arn,
          aws_iam_role.test3.arn,
          aws_iam_role.test5.arn,
          aws_iam_role.test2.arn,
        ]
      }
      Resource = [
        aws_s3_bucket.test.arn,
        "${aws_s3_bucket.test.arn}/*",
      ]
    }]
    Version = "2012-10-17"
  })
}
`, rName))
}

func testAccBucketPolicyIAMRoleOrderIAMPolicyDocNotPrincipalConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccBucketPolicyIAMRoleOrderBaseConfig(rName),
		`
data "aws_caller_identity" "current" {}
data "aws_iam_policy_document" "test" {
  statement {
    sid = "DenyInfected"
    actions = [
      "s3:GetObject",
      "s3:PutObjectTagging",
    ]
    effect = "Deny"
    not_principals {
      identifiers = [
        aws_iam_role.test2.arn,
        aws_iam_role.test3.arn,
        aws_iam_role.test4.arn,
        aws_iam_role.test1.arn,
        aws_iam_role.test5.arn,
        data.aws_caller_identity.current.arn,
      ]
      type = "AWS"
    }
    resources = [
      "${aws_s3_bucket.test.arn}/*",
    ]
    condition {
      test     = "StringEquals"
      variable = "s3:ExistingObjectTag/av-status"
      values   = ["INFECTED"]
    }
  }
}
resource "aws_s3_bucket_policy" "bucket" {
  bucket = aws_s3_bucket.test.bucket
  policy = data.aws_iam_policy_document.test.json
}
`)
}

func testAccBucketPolicy_Migrate_NoChangeConfig(bucketName, partition string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}
resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}
resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.id
  policy = %[2]s
}
`, bucketName, strconv.Quote(testAccBucketPolicy(bucketName, partition)))
}

func testAccBucketPolicy_Migrate_WithChangeConfig(bucketName, partition string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}
resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}
resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.id
  policy = %[2]s
}
`, bucketName, strconv.Quote(testAccBucketPolicyUpdated(bucketName, partition)))
}

func testAccBucketPolicyUpdated(bucketName, partition string) string {
>>>>>>> e0eb24f7 (feat(object): add support for bucket policy)
	return fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
<<<<<<< HEAD
        "SCW": "*"
      },
      "Action": "s3:PutObject",
      "Resource": "%s/*"
>>>>>>> c2713163 (Fix)
    }
    resources = [
      aws_s3_bucket.test.arn,
      "${aws_s3_bucket.test.arn}/*",
    ]
  }
}
resource "aws_s3_bucket_policy" "bucket" {
  bucket = aws_s3_bucket.test.bucket
  policy = data.aws_iam_policy_document.test.json
}
`, rName))
}

<<<<<<< HEAD
func testAccBucketPolicyIAMRoleOrderJSONEncodeConfig(rName string) string {
	return ConfigCompose(
		testAccBucketPolicyIAMRoleOrderBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket_policy" "bucket" {
  bucket = aws_s3_bucket.test.bucket
  policy = jsonencode({
    Id = %[1]q
    Statement = [{
      Action = [
        "s3:DeleteBucket",
        "s3:ListBucket",
        "s3:ListBucketVersions",
      ]
      Effect = "Allow"
      Principal = {
        AWS = [
          aws_iam_role.test2.arn,
          aws_iam_role.test1.arn,
          aws_iam_role.test4.arn,
          aws_iam_role.test3.arn,
          aws_iam_role.test5.arn,
        ]
      }
      Resource = [
        aws_s3_bucket.test.arn,
        "${aws_s3_bucket.test.arn}/*",
      ]
    }]
    Version = "2012-10-17"
  })
}
`, rName))
=======
func testAccCheckBucketPolicy(tt *TestTools, n string, policy string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		s3Client, err := newS3ClientFromMeta(tt.Meta)
		if err != nil {
			return err
		}

		out, err := s3Client.GetBucketPolicy(&s3.GetBucketPolicyInput{
			Bucket: aws.String(rs.Primary.ID),
		})

		if policy == "" {
			if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "NoSuchBucketPolicy" {
				// expected
				return nil
			}
			if err == nil {
				return fmt.Errorf("expected no policy, got: %#v", *out.Policy)
			}
			return fmt.Errorf("GetBucketPolicy error: %v, expected %s", err, policy)
		}
		if err != nil {
			return fmt.Errorf("GetBucketPolicy error: %v, expected %s", err, policy)
		}

		if v := out.Policy; v == nil {
			if policy != "" {
				return fmt.Errorf("bad policy, found nil, expected: %s", policy)
			}
		} else {
			expected := make(map[string]interface{})
			if err := json.Unmarshal([]byte(policy), &expected); err != nil {
				return err
			}
			actual := make(map[string]interface{})
			if err := json.Unmarshal([]byte(*v), &actual); err != nil {
				return err
			}

			if !reflect.DeepEqual(expected, actual) {
				return fmt.Errorf("bad policy, expected: %#v, got %#v", expected, actual)
			}
		}

		return nil
	}
>>>>>>> c2713163 (Fix)
}

func testAccBucketPolicyIAMRoleOrderJSONEncodeOrder2Config(rName string) string {
	return ConfigCompose(
		testAccBucketPolicyIAMRoleOrderBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket_policy" "bucket" {
  bucket = aws_s3_bucket.test.bucket
  policy = jsonencode({
    Id = %[1]q
    Statement = [{
      Action = [
        "s3:DeleteBucket",
        "s3:ListBucket",
        "s3:ListBucketVersions",
      ]
      Effect = "Allow"
      Principal = {
        AWS = [
          aws_iam_role.test2.arn,
          aws_iam_role.test3.arn,
          aws_iam_role.test5.arn,
          aws_iam_role.test1.arn,
          aws_iam_role.test4.arn,
        ]
      }
      Resource = [
        aws_s3_bucket.test.arn,
        "${aws_s3_bucket.test.arn}/*",
      ]
    }]
    Version = "2012-10-17"
  })
}
`, rName))
}

<<<<<<< HEAD
func testAccBucketPolicyIAMRoleOrderJSONEncodeOrder3Config(rName string) string {
	return ConfigCompose(
		testAccBucketPolicyIAMRoleOrderBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket_policy" "bucket" {
  bucket = aws_s3_bucket.test.bucket
  policy = jsonencode({
    Id = %[1]q
    Statement = [{
      Action = [
        "s3:DeleteBucket",
        "s3:ListBucket",
        "s3:ListBucketVersions",
      ]
      Effect = "Allow"
      Principal = {
        AWS = [
          aws_iam_role.test4.arn,
          aws_iam_role.test1.arn,
          aws_iam_role.test3.arn,
          aws_iam_role.test5.arn,
          aws_iam_role.test2.arn,
        ]
      }
      Resource = [
        aws_s3_bucket.test.arn,
        "${aws_s3_bucket.test.arn}/*",
      ]
    }]
    Version = "2012-10-17"
  })
}
`, rName))
}

func testAccBucketPolicyIAMRoleOrderIAMPolicyDocNotPrincipalConfig(rName string) string {
	return ConfigCompose(
		testAccBucketPolicyIAMRoleOrderBaseConfig(rName),
		`
data "aws_caller_identity" "current" {}
data "aws_iam_policy_document" "test" {
  statement {
    sid = "DenyInfected"
    actions = [
      "s3:GetObject",
      "s3:PutObjectTagging",
    ]
    effect = "Deny"
    not_principals {
      identifiers = [
        aws_iam_role.test2.arn,
        aws_iam_role.test3.arn,
        aws_iam_role.test4.arn,
        aws_iam_role.test1.arn,
        aws_iam_role.test5.arn,
        data.aws_caller_identity.current.arn,
      ]
      type = "AWS"
=======
func testAccBucketPolicy(bucketName string) string {
	return fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "SCW": "*"
      },
      "Action": "s3:GetObject",
      "Resource": "%[1]s/*"
>>>>>>> c2713163 (Fix)
    }
    resources = [
      "${aws_s3_bucket.test.arn}/*",
    ]
    condition {
      test     = "StringEquals"
      variable = "s3:ExistingObjectTag/av-status"
      values   = ["INFECTED"]
    }
  }
}
<<<<<<< HEAD
resource "aws_s3_bucket_policy" "bucket" {
  bucket = aws_s3_bucket.test.bucket
  policy = data.aws_iam_policy_document.test.json
=======

func testaccbucketconfigWithpolicy(bucketName string) string {
	return fmt.Sprintf(`
resource "scaleway_object_bucket" "test" {
  name = %[1]q
  acl    = "private"
  policy = %[2]s
>>>>>>> c2713163 (Fix)
}
<<<<<<< HEAD
`)
=======
`, bucketName, strconv.Quote(testAccBucketPolicy(bucketName)))
=======
        "AWS": "*"
      },
      "Action": "s3:PutObject",
      "Resource": "arn:%[1]s:s3:::%[2]s/*"
    }
  ]
}`, partition, bucketName)
>>>>>>> e0eb24f7 (feat(object): add support for bucket policy)
>>>>>>> 65b6efaa (feat(object): add support for bucket policy)
}
