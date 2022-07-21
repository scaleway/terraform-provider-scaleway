package scaleway

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccIAMPolicyDocumentDataSource_basic(t *testing.T) {
	// This really ought to be able to be a unit test rather than an
	// acceptance test, but just instantiating the AWS provider requires
	// some AWS API calls, and so this needs valid AWS credentials to work.
	tt := NewTestTools(t)
	defer tt.Cleanup()

	bucketName := "TestAccIAMPolicyDocumentDataSource_basic"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyDocumentConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaleway_object_policy_document.test", "json",
						testAccPolicyDocumentExpectedJSON(bucketName),
					),
				),
			},
		},
	})
}

func TestAccIAMPolicyDocumentDataSource_source(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resourceName := "TestAccIAMPolicyDocumentDataSource_source"
	// This really ought to be able to be a unit test rather than an
	// acceptance test, but just instantiating the AWS provider requires
	// some AWS API calls, and so this needs valid AWS credentials to work.
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyDocumentSourceConfigDeprecated,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaleway_object_policy_document.test_source", "json",
						testAccPolicyDocumentSourceExpectedJSON(resourceName),
					),
				),
			},
			{
				Config: testAccPolicyDocumentSourceBlankConfigDeprecated,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaleway_object_policy_document.test_source_blank", "json",
						testAccPolicyDocumentSourceBlankExpectedJSON,
					),
				),
			},
		},
	})
}

func TestAccIAMPolicyDocumentDataSource_sourceList(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyDocumentSourceListConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaleway_object_policy_document.test_source_list", "json",
						testAccPolicyDocumentSourceListExpectedJSON,
					),
				),
			},
		},
	})
}

func TestAccIAMPolicyDocumentDataSource_sourceConflicting(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyDocumentSourceConflictingConfigDeprecated,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaleway_object_policy_document.test_source_conflicting", "json",
						testAccPolicyDocumentSourceConflictingExpectedJSON,
					),
				),
			},
		},
	})
}

func TestAccIAMPolicyDocumentDataSource_sourceListConflicting(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccPolicyDocumentSourceListConflictingConfig,
				ExpectError: regexp.MustCompile(`duplicate Sid (.*?)`),
			},
		},
	})
}

func TestAccIAMPolicyDocumentDataSource_override(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyDocumentOverrideConfigDeprecated,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaleway_object_policy_document.test_override", "json",
						testAccPolicyDocumentOverrideExpectedJSON,
					),
				),
			},
		},
	})
}

func TestAccIAMPolicyDocumentDataSource_overrideList(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyDocumentOverrideListConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaleway_object_policy_document.test_override_list", "json",
						testAccPolicyDocumentOverrideListExpectedJSON,
					),
				),
			},
		},
	})
}

func TestAccIAMPolicyDocumentDataSource_noStatementMerge(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyDocumentNoStatementMergeConfigDeprecated,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaleway_object_policy_document.yak_politik", "json",
						testAccPolicyDocumentNoStatementMergeExpectedJSON,
					),
				),
			},
		},
	})
}

func TestAccIAMPolicyDocumentDataSource_noStatementOverride(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyDocumentNoStatementOverrideConfigDeprecated,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaleway_object_policy_document.yak_politik", "json",
						testAccPolicyDocumentNoStatementOverrideExpectedJSON,
					),
				),
			},
		},
	})
}

func TestAccIAMPolicyDocumentDataSource_duplicateSid(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccPolicyDocumentDuplicateSidConfig,
				ExpectError: regexp.MustCompile(`duplicate Sid`),
			},
			{
				Config: testAccPolicyDocumentDuplicateBlankSidConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaleway_object_policy_document.test", "json",
						testAccPolicyDocumentDuplicateBlankSidExpectedJSON,
					),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/10777
func TestAccIAMPolicyDocumentDataSource_StatementPrincipalIdentifiers_stringAndSlice(t *testing.T) {
	dataSourceName := "data.scaleway_object_policy_document.test"
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resourceName := "foobar"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyDocumentStatementPrincipalIdentifiersStringAndSliceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "json", testAccPolicyDocumentExpectedJSONStatementPrincipalIdentifiersStringAndSlice(resourceName)),
				),
			},
		},
	})
}

var testAccPolicyDocumentStatementPrincipalIdentifiersMultiplePrincipalsConfig = `
data "aws_partition" "current" {}

data "scaleway_object_policy_document" "test" {
  statement {
    actions   = ["*"]
    resources = ["*"]
    sid       = "StatementPrincipalIdentifiersStringAndSlice"

    principals {
      identifiers = [
        "arn:${data.aws_partition.current.partition}:iam::111111111111:root",
        "arn:${data.aws_partition.current.partition}:iam::222222222222:root",
      ]
      type = "AWS"
    }

    principals {
      identifiers = [
        "arn:${data.aws_partition.current.partition}:iam::333333333333:root",
      ]
      type = "AWS"
    }

    principals {
      identifiers = [
        "arn:${data.aws_partition.current.partition}:iam::444444444444:root",
      ]
      type = "AWS"
    }
  }
}
`

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/10777
func TestAccIAMPolicyDocumentDataSource_StatementPrincipalIdentifiers_multiplePrincipals(t *testing.T) {
	dataSourceName := "data.scaleway_object_policy_document.test"
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resourceName := "foobar"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyDocumentStatementPrincipalIdentifiersMultiplePrincipalsConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "json", testAccPolicyDocumentExpectedJSONStatementPrincipalIdentifiersMultiplePrincipals(resourceName)),
				),
			},
		},
	})
}

var testAccPolicyDocumentConfig = `
data "aws_partition" "current" {}

data "scaleway_object_policy_document" "test" {
  policy_id = "policy_id"

  statement {
    sid = "1"
    actions = [
      "s3:ListAllMyBuckets",
      "s3:GetBucketLocation",
    ]
    resources = [
      "arn:${data.aws_partition.current.partition}:s3:::*",
    ]
  }

  statement {
    actions = [
      "s3:ListBucket",
    ]

    resources = [
      "arn:${data.aws_partition.current.partition}:s3:::foo",
    ]

    condition {
      test     = "StringLike"
      variable = "s3:prefix"
      values = [
        "home/",
        "",
        "home/&{aws:username}/",
      ]
    }

    not_principals {
      type        = "AWS"
      identifiers = ["arn:blahblah:example"]
    }
  }

  statement {
    actions = [
      "s3:*",
    ]

    resources = [
      "arn:${data.aws_partition.current.partition}:s3:::foo/home/&{aws:username}",
      "arn:${data.aws_partition.current.partition}:s3:::foo/home/&{aws:username}/*",
    ]

    principals {
      type        = "AWS"
      identifiers = ["arn:blahblah:example"]
    }
  }

  statement {
    effect        = "Deny"
    not_actions   = ["s3:*"]
    not_resources = ["arn:${data.aws_partition.current.partition}:s3:::*"]
  }

  # Normalization of wildcard principals

  statement {
    effect  = "Allow"
    actions = ["kinesis:*"]

    principals {
      type        = "AWS"
      identifiers = ["*"]
    }
  }

  statement {
    effect  = "Allow"
    actions = ["firehose:*"]

    principals {
      type        = "*"
      identifiers = ["*"]
    }
  }
}
`

func testAccPolicyDocumentExpectedJSON(rname string) string {
	return fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Id": "policy_id",
  "Statement": [
    {
      "Sid": "1",
      "Effect": "Allow",
      "Action": [
        "s3:ListAllMyBuckets",
        "s3:GetBucketLocation"
      ],
      "Resource": "%[1]s"
    },
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": "s3:ListBucket",
      "Resource": "%[1]s",
      "NotPrincipal": {
        "AWS": "arn:blahblah:example"
      },
      "Condition": {
        "StringLike": {
          "s3:prefix": [
            "home/",
            "",
            "home/${aws:username}/"
          ]
        }
      }
    },
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": "s3:*",
      "Resource": [
        "%[1]s:s3:::foo/home/${aws:username}/*",
        "%[1]s:s3:::foo/home/${aws:username}"
      ],
      "Principal": {
        "AWS": "arn:blahblah:example"
      }
    },
    {
      "Sid": "",
      "Effect": "Deny",
      "NotAction": "s3:*",
      "NotResource": "arn:%[1]s:s3:::*"
    }
  ]
}`, rname)
}

var testAccPolicyDocumentSourceConfigDeprecated = `
data "aws_partition" "current" {}

data "scaleway_object_policy_document" "test" {
  policy_id = "policy_id"

  statement {
    sid = "1"
    actions = [
      "s3:ListAllMyBuckets",
      "s3:GetBucketLocation",
    ]
    resources = [
      "arn:${data.aws_partition.current.partition}:s3:::*",
    ]
  }

  statement {
    actions = [
      "s3:ListBucket",
    ]

    resources = [
      "arn:${data.aws_partition.current.partition}:s3:::foo",
    ]

    condition {
      test     = "StringLike"
      variable = "s3:prefix"
      values = [
        "home/",
        "home/&{aws:username}/",
      ]
    }

    not_principals {
      type        = "AWS"
      identifiers = ["arn:blahblah:example"]
    }
  }

  statement {
    actions = [
      "s3:*",
    ]

    resources = [
      "arn:${data.aws_partition.current.partition}:s3:::foo/home/&{aws:username}",
      "arn:${data.aws_partition.current.partition}:s3:::foo/home/&{aws:username}/*",
    ]

    principals {
      type = "AWS"
      identifiers = [
        "arn:blahblah:example",
        "arn:blahblahblah:example",
      ]
    }
  }

  statement {
    effect        = "Deny"
    not_actions   = ["s3:*"]
    not_resources = ["arn:${data.aws_partition.current.partition}:s3:::*"]
  }

  # Normalization of wildcard principals

  statement {
    effect  = "Allow"
    actions = ["kinesis:*"]

    principals {
      type        = "AWS"
      identifiers = ["*"]
    }
  }

  statement {
    effect  = "Allow"
    actions = ["firehose:*"]

    principals {
      type        = "*"
      identifiers = ["*"]
    }
  }
}

data "scaleway_object_policy_document" "test_source" {
  source_json = data.scaleway_object_policy_document.test.json

  statement {
    sid       = "SourceJSONTest1"
    actions   = ["*"]
    resources = ["*"]
  }
}
`

func testAccPolicyDocumentSourceExpectedJSON(resourceName string) string {
	return fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Id": "policy_id",
  "Statement": [
    {
      "Sid": "1",
      "Effect": "Allow",
      "Action": [
        "s3:ListAllMyBuckets",
        "s3:GetBucketLocation"
      ],
      "Resource": "arn:%[1]s:s3:::*"
    },
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": "s3:ListBucket",
      "Resource": "arn:%[1]s:s3:::foo",
      "NotPrincipal": {
        "AWS": "arn:blahblah:example"
      },
      "Condition": {
        "StringLike": {
          "s3:prefix": [
            "home/",
            "home/${aws:username}/"
          ]
        }
      }
    },
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": "s3:*",
      "Resource": [
        "arn:%[1]s:s3:::foo/home/${aws:username}/*",
        "arn:%[1]s:s3:::foo/home/${aws:username}"
      ],
      "Principal": {
        "AWS": [
          "arn:blahblahblah:example",
          "arn:blahblah:example"
        ]
      }
    },
    {
      "Sid": "",
      "Effect": "Deny",
      "NotAction": "s3:*",
      "NotResource": "arn:%[1]s:s3:::*"
    },
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": "kinesis:*",
      "Principal": {
        "SCW": "*"
      }
    },
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": "firehose:*",
      "Principal": "*"
    },
    {
      "Sid": "SourceJSONTest1",
      "Effect": "Allow",
      "Action": "*",
      "Resource": "*"
    }
  ]
}`, resourceName)
}

var testAccPolicyDocumentSourceListConfig = `
data "scaleway_object_policy_document" "policy_a" {
  statement {
    sid     = ""
    effect  = "Allow"
    actions = ["foo:ActionOne"]
  }

  statement {
    sid     = "validSidOne"
    effect  = "Allow"
    actions = ["bar:ActionOne"]
  }
}

data "scaleway_object_policy_document" "policy_b" {
  statement {
    sid     = "validSidTwo"
    effect  = "Deny"
    actions = ["foo:ActionTwo"]
  }
}

data "scaleway_object_policy_document" "policy_c" {
  statement {
    sid     = ""
    effect  = "Allow"
    actions = ["bar:ActionTwo"]
  }
}

data "scaleway_object_policy_document" "test_source_list" {
  version = "2012-10-17"

  source_policy_documents = [
    data.scaleway_object_policy_document.policy_a.json,
    data.scaleway_object_policy_document.policy_b.json,
    data.scaleway_object_policy_document.policy_c.json
  ]
}
`

var testAccPolicyDocumentSourceListExpectedJSON = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": "foo:ActionOne"
    },
    {
      "Sid": "validSidOne",
      "Effect": "Allow",
      "Action": "bar:ActionOne"
    },
    {
      "Sid": "validSidTwo",
      "Effect": "Deny",
      "Action": "foo:ActionTwo"
    },
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": "bar:ActionTwo"
    }
  ]
}`

var testAccPolicyDocumentSourceBlankConfigDeprecated = `
data "scaleway_object_policy_document" "test_source_blank" {
  source_json = ""

  statement {
    sid       = "SourceJSONTest2"
    actions   = ["*"]
    resources = ["*"]
  }
}
`

var testAccPolicyDocumentSourceBlankExpectedJSON = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "SourceJSONTest2",
      "Effect": "Allow",
      "Action": "*",
      "Resource": "*"
    }
  ]
}`

var testAccPolicyDocumentSourceConflictingConfigDeprecated = `
data "scaleway_object_policy_document" "test_source" {
  statement {
    sid       = "SourceJSONTestConflicting"
    actions   = ["iam:*"]
    resources = ["*"]
  }
}

data "scaleway_object_policy_document" "test_source_conflicting" {
  source_json = data.scaleway_object_policy_document.test_source.json

  statement {
    sid       = "SourceJSONTestConflicting"
    actions   = ["*"]
    resources = ["*"]
  }
}
`

var testAccPolicyDocumentSourceConflictingExpectedJSON = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "SourceJSONTestConflicting",
      "Effect": "Allow",
      "Action": "*",
      "Resource": "*"
    }
  ]
}`

var testAccPolicyDocumentSourceListConflictingConfig = `
data "scaleway_object_policy_document" "policy_a" {
  statement {
    sid     = ""
    effect  = "Allow"
    actions = ["foo:ActionOne"]
  }

  statement {
    sid     = "conflictSid"
    effect  = "Allow"
    actions = ["bar:ActionOne"]
  }
}

data "scaleway_object_policy_document" "policy_b" {
  statement {
    sid     = "validSid"
    effect  = "Deny"
    actions = ["foo:ActionTwo"]
  }
}

data "scaleway_object_policy_document" "policy_c" {
  statement {
    sid     = "conflictSid"
    effect  = "Allow"
    actions = ["bar:ActionTwo"]
  }
}

data "scaleway_object_policy_document" "test_source_list_conflicting" {
  version = "2012-10-17"

  source_policy_documents = [
    data.scaleway_object_policy_document.policy_a.json,
    data.scaleway_object_policy_document.policy_b.json,
    data.scaleway_object_policy_document.policy_c.json
  ]
}
`

var testAccPolicyDocumentOverrideConfigDeprecated = `
data "aws_partition" "current" {}

data "scaleway_object_policy_document" "override" {
  statement {
    sid = "SidToOverwrite"

    actions   = ["s3:*"]
    resources = ["*"]
  }
}

data "scaleway_object_policy_document" "test_override" {
  override_json = data.scaleway_object_policy_document.override.json

  statement {
    actions   = ["ec2:*"]
    resources = ["*"]
  }

  statement {
    sid = "SidToOverwrite"

    actions = ["s3:*"]

    resources = [
      "arn:${data.aws_partition.current.partition}:s3:::somebucket",
      "arn:${data.aws_partition.current.partition}:s3:::somebucket/*",
    ]
  }
}
`

var testAccPolicyDocumentOverrideExpectedJSON = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": "ec2:*",
      "Resource": "*"
    },
    {
      "Sid": "SidToOverwrite",
      "Effect": "Allow",
      "Action": "s3:*",
      "Resource": "*"
    }
  ]
}`

var testAccPolicyDocumentOverrideListConfig = `
data "scaleway_object_policy_document" "policy_a" {
  statement {
    sid     = ""
    effect  = "Allow"
    actions = ["foo:ActionOne"]
  }

  statement {
    sid     = "overrideSid"
    effect  = "Allow"
    actions = ["bar:ActionOne"]
  }
}

data "scaleway_object_policy_document" "policy_b" {
  statement {
    sid     = "validSid"
    effect  = "Deny"
    actions = ["foo:ActionTwo"]
  }
}

data "scaleway_object_policy_document" "policy_c" {
  statement {
    sid     = "overrideSid"
    effect  = "Deny"
    actions = ["bar:ActionOne"]
  }
}

data "scaleway_object_policy_document" "test_override_list" {
  version = "2012-10-17"

  override_policy_documents = [
    data.scaleway_object_policy_document.policy_a.json,
    data.scaleway_object_policy_document.policy_b.json,
    data.scaleway_object_policy_document.policy_c.json
  ]
}
`

var testAccPolicyDocumentOverrideListExpectedJSON = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": "foo:ActionOne"
    },
    {
      "Sid": "overrideSid",
      "Effect": "Deny",
      "Action": "bar:ActionOne"
    },
    {
      "Sid": "validSid",
      "Effect": "Deny",
      "Action": "foo:ActionTwo"
    }
  ]
}`

var testAccPolicyDocumentNoStatementMergeConfigDeprecated = `
data "scaleway_object_policy_document" "source" {
  statement {
    sid       = ""
    actions   = ["ec2:DescribeAccountAttributes"]
    resources = ["*"]
  }
}

data "scaleway_object_policy_document" "override" {
  statement {
    sid       = "OverridePlaceholder"
    actions   = ["s3:GetObject"]
    resources = ["*"]
  }
}

data "scaleway_object_policy_document" "yak_politik" {
  source_json   = data.scaleway_object_policy_document.source.json
  override_json = data.scaleway_object_policy_document.override.json
}
`

var testAccPolicyDocumentNoStatementMergeExpectedJSON = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": "ec2:DescribeAccountAttributes",
      "Resource": "*"
    },
    {
      "Sid": "OverridePlaceholder",
      "Effect": "Allow",
      "Action": "s3:GetObject",
      "Resource": "*"
    }
  ]
}`

var testAccPolicyDocumentNoStatementOverrideConfigDeprecated = `
data "scaleway_object_policy_document" "source" {
  statement {
    sid       = "OverridePlaceholder"
    actions   = ["ec2:DescribeAccountAttributes"]
    resources = ["*"]
  }
}

data "scaleway_object_policy_document" "override" {
  statement {
    sid       = "OverridePlaceholder"
    actions   = ["s3:GetObject"]
    resources = ["*"]
  }
}

data "scaleway_object_policy_document" "yak_politik" {
  source_json   = data.scaleway_object_policy_document.source.json
  override_json = data.scaleway_object_policy_document.override.json
}
`

var testAccPolicyDocumentNoStatementOverrideExpectedJSON = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "OverridePlaceholder",
      "Effect": "Allow",
      "Action": "s3:GetObject",
      "Resource": "*"
    }
  ]
}`

var testAccPolicyDocumentDuplicateSidConfig = `
data "scaleway_object_policy_document" "test" {
  statement {
    sid       = "1"
    effect    = "Allow"
    actions   = ["ec2:DescribeAccountAttributes"]
    resources = ["*"]
  }

  statement {
    sid       = "1"
    effect    = "Allow"
    actions   = ["s3:GetObject"]
    resources = ["*"]
  }
}
`

var testAccPolicyDocumentDuplicateBlankSidConfig = `
data "scaleway_object_policy_document" "test" {
  statement {
    sid       = ""
    effect    = "Allow"
    actions   = ["ec2:DescribeAccountAttributes"]
    resources = ["*"]
  }

  statement {
    sid       = ""
    effect    = "Allow"
    actions   = ["s3:GetObject"]
    resources = ["*"]
  }
}
`

var testAccPolicyDocumentDuplicateBlankSidExpectedJSON = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": "ec2:DescribeAccountAttributes",
      "Resource": "*"
    },
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": "s3:GetObject",
      "Resource": "*"
    }
  ]
}`

var testAccPolicyDocumentStatementPrincipalIdentifiersStringAndSliceConfig = `
data "aws_partition" "current" {}

data "scaleway_object_policy_document" "test" {
  statement {
    actions   = ["*"]
    resources = ["*"]
    sid       = "StatementPrincipalIdentifiersStringAndSlice"

    principals {
      identifiers = ["arn:${data.aws_partition.current.partition}:iam::111111111111:root"]
      type        = "AWS"
    }

    principals {
      identifiers = ["arn:${data.aws_partition.current.partition}:iam::222222222222:root", "arn:${data.aws_partition.current.partition}:iam::333333333333:root"]
      type        = "AWS"
    }
  }
}
`

func testAccPolicyDocumentExpectedJSONStatementPrincipalIdentifiersStringAndSlice(resourceName string) string {
	return fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "StatementPrincipalIdentifiersStringAndSlice",
      "Effect": "Allow",
      "Action": "*",
      "Resource": "*",
      "Principal": {
        "AWS": [
          "arn:%[1]s:iam::111111111111:root",
          "arn:%[1]s:iam::333333333333:root",
          "arn:%[1]s:iam::222222222222:root"
        ]
      }
    }
  ]
}`, resourceName)
}

func testAccPolicyDocumentExpectedJSONStatementPrincipalIdentifiersMultiplePrincipals(resourceName string) string {
	return fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "StatementPrincipalIdentifiersStringAndSlice",
      "Effect": "Allow",
      "Action": "*",
      "Resource": "*",
      "Principal": {
        "AWS": [
          "arn:%[1]s:iam::333333333333:root",
          "arn:%[1]s:iam::444444444444:root",
          "arn:%[1]s:iam::222222222222:root",
          "arn:%[1]s:iam::111111111111:root"
        ]
      }
    }
  ]
}`, resourceName)
}

// CheckResourceAttrEquivalentJSON is a TestCheckFunc that compares a JSON value with an expected value. Both JSON
// values are normalized before being compared.
func CheckResourceAttrEquivalentJSON(resourceName, attributeName, expectedJSON string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		is, err := PrimaryInstanceState(s, resourceName)
		if err != nil {
			return err
		}

		v, ok := is.Attributes[attributeName]
		if !ok {
			return fmt.Errorf("%s: No attribute %q found", resourceName, attributeName)
		}

		vNormal, err := structure.NormalizeJsonString(v)
		if err != nil {
			return fmt.Errorf("%s: Error normalizing JSON in %q: %w", resourceName, attributeName, err)
		}

		expectedNormal, err := structure.NormalizeJsonString(expectedJSON)
		if err != nil {
			return fmt.Errorf("error normalizing expected JSON: %w", err)
		}

		if vNormal != expectedNormal {
			return fmt.Errorf("%s: Attribute %q expected\n%s\ngot\n%s", resourceName, attributeName, expectedJSON, v)
		}
		return nil
	}
}

// Copied and inlined from the SDK testing code
func PrimaryInstanceState(s *terraform.State, name string) (*terraform.InstanceState, error) {
	rs, ok := s.RootModule().Resources[name]
	if !ok {
		return nil, fmt.Errorf("not found: %s", name)
	}

	is := rs.Primary
	if is == nil {
		return nil, fmt.Errorf("no primary instance: %s", name)
	}

	return is, nil
}
