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
	tt := NewTestTools(t)
	defer tt.Cleanup()

	bucketName := "test-acc-iam-policy-document-data-source-basic"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					data "scaleway_object_policy_document" "test" {
						policy_id = "policy_id"
					
						statement {
							sid = "1"
							actions = [
						  		"s3:ListBucket",
						  		"s3:GetBucketWebsite",
							]
							resources = [
						  		%[1]q,
							]
					  	}
					
					  	statement {
							actions = [
						  		"s3:ListBucket",
							]
					
							resources = [
						  		%[1]q,
							]
					
							condition {
						  		test     = "StringLike"
						  		variable = "aws:SourceIp"
						  		values = [
									"home/",
									"",
									"home/&{aws:username}/",
						  		]
							}
					  	}
					
						statement {
							actions = [
						  		"s3:*",
							]
					
							resources = [
						  		"%[1]s",
						  		"%[1]s/*",
							]
					
							principals {
						  		type        = "SCW"
						  		identifiers = ["arn:blahblah:example"]
							}
					  	}
					
					  	statement {
							effect        = "Deny"
							not_actions   = ["s3:*"]
							not_resources = [%[1]q]
					  	}
					}
				`, bucketName),
				Check: resource.ComposeTestCheckFunc(
					CheckResourceAttrEquivalentJSON("data.scaleway_object_policy_document.test", "json", fmt.Sprintf(`{
					  "Version": "2012-10-17",
					  "Id": "policy_id",
					  "Statement": [
						{
						  "Sid": "1",
						  "Effect": "Allow",
						  "Action": [
							"s3:ListBucket",
							"s3:GetBucketWebsite"
						  ],
						  "Resource": "%[1]s"
						},
						{
						  "Sid": "",
						  "Effect": "Allow",
						  "Action": "s3:ListBucket",
						  "Resource": "%[1]s",
						  "Condition": [
							{
							  "Test": "StringLike",
							  "Variable": "aws:SourceIp",
							  "Values": [
								"home/",
								"",
								"home/${aws:username}/"
							  ]
							}
						  ]
						},
						{
						  "Sid": "",
						  "Effect": "Allow",
						  "Action": "s3:*",
						  "Resource": [
							"%[1]s/*",
							"%[1]s"
						  ],
						  "Principal": [
							{
							  "Type": "SCW",
							  "Identifiers": "arn:blahblah:example"
							}
						  ]
						},
						{
						  "Sid": "",
						  "Effect": "Deny",
						  "NotAction": "s3:*",
						  "NotResource": "%[1]s"
						}
					  ]
					}`, bucketName)),
				),
			},
		},
	})
}

func TestAccIAMPolicyDocumentDataSource_source(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resourceName := "TestAccIAMPolicyDocumentDataSource_source"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					data "scaleway_object_policy_document" "test" {
						policy_id = "policy_id"
					
						statement {
							sid = "1"
							actions = [
						  		"s3:DeleteObjectVersionTagging",
						  		"s3:DeleteObjectTagging",
							]
							
							resources = [
						  		"%[1]s",
							]
					  	}
					
						statement {
							actions = [
						  		"s3:ListBucket",
							]
					
							resources = [
						  		"%[1]s",
							]
					
							condition {
						  		test     = "StringLike"
						  		variable = "aws:SourceIp"
						  		values = [
									"1.2.3.4",
									"2.3.4.5",
						  		]
							}
					
							not_principals {
						  		type        = "SCW"
						  		identifiers = ["arn:blahblah:example"]
							}
					  	}
					
						statement {
							actions = [
						  		"s3:*",
							]
					
							resources = [
								"%[1]s",
								"%[1]s/*",
							]
					
							principals {
						  		type = "SCW"
						  		identifiers = [
									"project_id:11111111-1111-1111-1111-111111111111",
									"project_id:11111111-1111-1111-1111-111111111111",
						  		]
							}
						}
					
					  	statement {
							effect        = "Deny"
							not_actions   = ["s3:*"]
							not_resources = ["%[1]s"]
					  	}
					
						# Normalization of wildcard principals
					  	statement {
							effect  = "Allow"
							actions = ["s3:*"]
					
							principals {
						  		type        = "SCW"
						  		identifiers = ["*"]
							}
					  	}
					
						statement {
							effect  = "Allow"
							actions = ["s3:*"]
					
							principals {
								type        = "*"
								identifiers = ["*"]
							}
					  	}
					}
					
					data "scaleway_object_policy_document" "test_source" {
						source_policy_documents = [data.scaleway_object_policy_document.test.json]
					
						statement {
							sid       = "SourceJSONTest1"
							actions   = ["*"]
							resources = ["*"]
					  	}
					}`, resourceName),
				Check: resource.ComposeTestCheckFunc(
					CheckResourceAttrEquivalentJSON("data.scaleway_object_policy_document.test_source", "json",
						fmt.Sprintf(`{
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
							  "Resource": "%[1]s",
							  "NotPrincipal": {
								"SCW": "project_id:"
							  },
							  "Condition": {
								"StringLike": {
								  "aws:SourceIp": [
									"1.2.3.4",
									"2.3.4.5"
								  ]
								}
							  }
							},
							{
							  "Sid": "",
							  "Effect": "Allow",
							  "Action": "s3:*",
							  "Resource": [
								"%[1]s/*",
								"%[1]s"
							  ],
							  "Principal": {
								"SCW": [
									"project_id:11111111-1111-1111-1111-111111111111",
									"project_id:11111111-1111-1111-1111-111111111111"
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
							  "Sid": "SourceJSONTest1",
							  "Effect": "Allow",
							  "Action": "*",
							  "Resource": "*"
							}
						  ]
						}`, resourceName),
					),
				),
			},
			{
				Config: `
					data "scaleway_object_policy_document" "test_source_blank" {
						source_policy_documents = [""]
					
						statement {
							sid       = "SourceJSONTest2"
							actions   = ["*"]
							resources = ["*"]
					  	}
					}`,
				Check: resource.ComposeTestCheckFunc(
					CheckResourceAttrEquivalentJSON("data.scaleway_object_policy_document.test_source_blank", "json",
						`{
							"Version": "2012-10-17",
							"Statement": [
								{
									"Sid": "SourceJSONTest2",
									"Effect": "Allow",
									"Action": "*",
									"Resource": "*"
								}
							]
						}`,
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
				Config: `
					data "scaleway_object_policy_document" "policy_a" {
						statement {
							sid     = ""
							effect  = "Allow"
							actions = ["s3:AbortMultipartUpload"]
					  	}
					
					  	statement {
							sid     = "validSidOne"
							effect  = "Allow"
							actions = ["s3:DeleteBucketWebsite"]
					  	}
					}
					
					data "scaleway_object_policy_document" "policy_b" {
						statement {
							sid     = "validSidTwo"
							effect  = "Deny"
							actions = ["s3:DeleteObject"]
					  	}
					}

					data "scaleway_object_policy_document" "policy_c" {
  						statement {
						    sid     = ""
    						effect  = "Allow"
    						actions = ["s3:DeleteObjectTagging"]
  						}
					}

					data "scaleway_object_policy_document" "test_source_list" {
  						version = "2012-10-17"

  						source_policy_documents = [
    						data.scaleway_object_policy_document.policy_a.json,
    						data.scaleway_object_policy_document.policy_b.json,
    						data.scaleway_object_policy_document.policy_c.json
  						]
					}`,
				Check: resource.ComposeTestCheckFunc(
					CheckResourceAttrEquivalentJSON("data.scaleway_object_policy_document.test_source_list", "json",
						`{
						  "Version": "2012-10-17",
						  "Statement": [
							{
							  "Sid": "",
							  "Effect": "Allow",
							  "Action": "s3:AbortMultipartUpload"
							},
							{
							  "Sid": "validSidOne",
							  "Effect": "Allow",
							  "Action": "s3:DeleteBucketWebsite"
							},
							{
							  "Sid": "validSidTwo",
							  "Effect": "Deny",
							  "Action": "s3:DeleteObject"
							},
							{
							  "Sid": "",
							  "Effect": "Allow",
							  "Action": "s3:DeleteObjectTagging"
							}
						  ]
						}`,
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
				Config: `
					data "scaleway_object_policy_document" "test_source" { 
						statement {
							sid       = "SourceJSONTestConflicting"
							actions   = ["s3:*"]
							resources = ["*"]
						}
					}

					data "scaleway_object_policy_document" "test_source_conflicting" {
  						source_policy_documents = [data.scaleway_object_policy_document.test_source.json]

						statement {
							sid       = "SourceJSONTestConflicting"
							actions   = ["*"]
							resources = ["*"]
						}
					}`,
				Check: resource.ComposeTestCheckFunc(
					CheckResourceAttrEquivalentJSON("data.scaleway_object_policy_document.test_source_conflicting", "json",
						`{
						  "Version": "2012-10-17",
						  "Statement": [
							{
							  "Sid": "SourceJSONTestConflicting",
							  "Effect": "Allow",
							  "Action": "*",
							  "Resource": "*"
							}
						  ]
						}`,
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
				Config: `
					data "scaleway_object_policy_document" "policy_a" {
						statement {
							sid     = ""
    						effect  = "Allow"
    						actions = ["s3:AbortMultipartUpload"]
  						}

						statement {
							sid     = "conflictSid"
							effect  = "Allow"
							actions = ["s3:DeleteBucketWebsite"]
						}
					}

					data "scaleway_object_policy_document" "policy_b" {
						statement {
							sid     = "validSid"
							effect  = "Deny"
							actions = ["s3:DeleteObject"]
					  	}
					}

					data "scaleway_object_policy_document" "policy_c" {
  						statement {
    						sid     = "conflictSid"
    						effect  = "Allow"
    						actions = ["s3:DeleteObjectTagging"]
  						}
					}

					data "scaleway_object_policy_document" "test_source_list_conflicting" {
  						version = "2012-10-17"

  						source_policy_documents = [
    						data.scaleway_object_policy_document.policy_a.json,
    						data.scaleway_object_policy_document.policy_b.json,
    						data.scaleway_object_policy_document.policy_c.json
  						]
					}`,
				ExpectError: regexp.MustCompile(`duplicate Sid (.*?)`),
			},
		},
	})
}

func TestAccIAMPolicyDocumentDataSource_override(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	bucketName := "test-acc-iam-policy-document-data-source-override"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					data "scaleway_object_policy_document" "override" {
  						statement {
    						sid = "SidToOverwrite"
    						actions   = ["s3:*"]
    						resources = ["*"]
  						}
					}

					data "scaleway_object_policy_document" "test_override" {
  						override_policy_documents = [data.scaleway_object_policy_document.override.json]

						statement {
							actions   = ["s3:*"]
							resources = ["*"]
						}

						statement {
							sid = "SidToOverwrite"
    						actions = ["s3:*"]
    						resources = [
      							"%[1]s",
      							"%[1]s/*",
    						]
  						}
					}`, bucketName),
				Check: resource.ComposeTestCheckFunc(
					CheckResourceAttrEquivalentJSON("data.scaleway_object_policy_document.test_override", "json",
						`{
						  "Version": "2012-10-17",
						  "Statement": [
							{
							  "Sid": "",
							  "Effect": "Allow",
							  "Action": "s3:*",
							  "Resource": "*"
							},
							{
							  "Sid": "SidToOverwrite",
							  "Effect": "Allow",
							  "Action": "s3:*",
							  "Resource": "*"
							}
						  ]
						}`,
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
				Config: `
					data "scaleway_object_policy_document" "policy_a" {
  						statement {
							sid     = ""
							effect  = "Allow"
							actions = ["s3:AbortMultipartUpload"]
  						}

						statement {
    						sid     = "overrideSid"
    						effect  = "Allow"
    						actions = ["s3:AbortMultipartUpload"]
  						}
					}

					data "scaleway_object_policy_document" "policy_b" {
						statement {
							sid     = "validSid"
							effect  = "Deny"
							actions = ["s3:DeleteBucketWebsite"]
					  	}
					}

					data "scaleway_object_policy_document" "policy_c" {
  						statement {
    						sid     = "overrideSid"
    						effect  = "Deny"
    						actions = ["s3:AbortMultipartUpload"]
  						}
					}

					data "scaleway_object_policy_document" "test_override_list" {
  						version = "2012-10-17"

  						override_policy_documents = [
    						data.scaleway_object_policy_document.policy_a.json,
    						data.scaleway_object_policy_document.policy_b.json,
    						data.scaleway_object_policy_document.policy_c.json
  						]
					}`,
				Check: resource.ComposeTestCheckFunc(
					CheckResourceAttrEquivalentJSON("data.scaleway_object_policy_document.test_override_list", "json",
						`{
							"Version": "2012-10-17",
							"Statement": [
								{
									"Sid": "",
									"Effect": "Allow",
									"Action": "s3:AbortMultipartUpload"
								},
								{
								  "Sid": "overrideSid",
								  "Effect": "Deny",
								  "Action": "s3:AbortMultipartUpload"
								},
								{
								  "Sid": "validSid",
								  "Effect": "Deny",
								  "Action": "s3:DeleteBucketWebsite"
								}
							  ]
							}`,
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
				Config: `
					data "scaleway_object_policy_document" "source" {
  						statement {
    						sid       = ""
    						actions   = ["s3:AbortMultipartUpload"]
    						resources = ["*"]
  						}
					}

					data "scaleway_object_policy_document" "override" {
  						statement {
    						sid       = "OverridePlaceholder"
    						actions   = ["s3:DeleteBucketWebsite"]
    						resources = ["*"]
  						}
					}

					data "scaleway_object_policy_document" "yak_politik" {
  						source_policy_documents   = [data.scaleway_object_policy_document.source.json]
  						override_policy_documents = [data.scaleway_object_policy_document.override.json]
					}`,
				Check: resource.ComposeTestCheckFunc(
					CheckResourceAttrEquivalentJSON("data.scaleway_object_policy_document.yak_politik", "json",
						`{
						  "Version": "2012-10-17",
						  "Statement": [
							{
							  "Sid": "",
							  "Effect": "Allow",
							  "Action": "s3:AbortMultipartUpload",
							  "Resource": "*"
							},
							{
							  "Sid": "OverridePlaceholder",
							  "Effect": "Allow",
							  "Action": "s3:DeleteBucketWebsite",
							  "Resource": "*"
							}
						  ]
						}`,
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
				Config: `
					data "scaleway_object_policy_document" "test" {
						statement {
							sid       = "1"
							effect    = "Allow"
							actions   = ["s3:GetObjectVersionTagging"]
							resources = ["*"]
					  	}
					
					  	statement {
							sid       = "1"
							effect    = "Allow"
							actions   = ["s3:GetObject"]
							resources = ["*"]
					  	}
					}`,
				ExpectError: regexp.MustCompile(`duplicate Sid`),
			},
			{
				Config: `
					data "scaleway_object_policy_document" "test" {
						statement {
							sid       = ""
							effect    = "Allow"
							actions   = ["s3:GetObjectVersionTagging"]
							resources = ["*"]
					  	}
					
					  	statement {
							sid       = ""
							effect    = "Allow"
							actions   = ["s3:GetObject"]
							resources = ["*"]
					  	}
					}`,
				Check: resource.ComposeTestCheckFunc(
					CheckResourceAttrEquivalentJSON("data.scaleway_object_policy_document.test", "json",
						`{
						  "Version": "2012-10-17",
						  "Statement": [
							{
							  "Sid": "",
							  "Effect": "Allow",
							  "Action": "s3:GetObjectVersionTagging",
							  "Resource": "*"
							},
							{
							  "Sid": "",
							  "Effect": "Allow",
							  "Action": "s3:GetObject",
							  "Resource": "*"
							}
						  ]
						}`,
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "scaleway_object_policy_document" "test" {
						statement {
							actions   = ["*"]
							resources = ["*"]
							sid       = "StatementPrincipalIdentifiersStringAndSlice"
					
							principals {
						  		identifiers = ["project_id:11111111-1111-1111-1111-111111111111"]
						  		type        = "SCW"
							}
					
							principals {
						  		identifiers = [
									"project_id:22222222-2222-2222-2222-222222222222", 
									"project_id:33333333-3333-3333-3333-333333333333"
								]
						  		type        = "SCW"
							}
					  	}
					}`,
				Check: resource.ComposeTestCheckFunc(
					CheckResourceAttrEquivalentJSON(dataSourceName, "json", `{
					  "Version": "2012-10-17",
					  "Statement": [
						{
						  "Sid": "StatementPrincipalIdentifiersStringAndSlice",
						  "Effect": "Allow",
						  "Action": "*",
						  "Resource": "*",
						  "Principal": [
							{
							  "Type": "SCW",
							  "Identifiers": "project_id:11111111-1111-1111-1111-111111111111"
							},
							{
							  "Type": "SCW",
							  "Identifiers": [
								"project_id:33333333-3333-3333-3333-333333333333",
								"project_id:22222222-2222-2222-2222-222222222222"
							  ]
							}
						  ]
						}
					  ]
					}`),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/10777
func TestAccIAMPolicyDocumentDataSource_StatementPrincipalIdentifiers_multiplePrincipals(t *testing.T) {
	dataSourceName := "data.scaleway_object_policy_document.test"
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "scaleway_object_policy_document" "test" {
						statement {
							actions   = ["*"]
							resources = ["*"]
							sid       = "StatementPrincipalIdentifiersStringAndSlice"
					
						principals {
							identifiers = [
								"project_id:44444444-4444-4444-4444-444444444444",
								"project_id:33333333-3333-3333-3333-333333333333",
						  	]
						  	type = "SCW"
						}
					
						principals {
							identifiers = [
								"project_id:22222222-2222-2222-2222-222222222222",
						  	]
							type = "SCW"
						}
					
						principals {
							identifiers = [
								"project_id:11111111-1111-1111-1111-111111111111",
						  	]
							type = "SCW"
						}
					  }
					}`,
				Check: resource.ComposeTestCheckFunc(
					CheckResourceAttrEquivalentJSON(dataSourceName, "json", `{
					  "Version": "2012-10-17",
					  "Statement": [
						{
						  "Sid": "StatementPrincipalIdentifiersStringAndSlice",
						  "Effect": "Allow",
						  "Action": "*",
						  "Resource": "*",
						  "Principal": [
							{
							  "Type": "SCW",
							  "Identifiers": "project_id:22222222-2222-2222-2222-222222222222"
							},
							{
							  "Type": "SCW",
							  "Identifiers": "project_id:11111111-1111-1111-1111-111111111111"
							},
							{
							  "Type": "SCW",
							  "Identifiers": [
								"project_id:44444444-4444-4444-4444-444444444444",
								"project_id:33333333-3333-3333-3333-333333333333"
							  ]
							}
						]
						}
					  ]
					}`),
				),
			},
		},
	})
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
