package scaleway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var dataSourcePolicyDocumentVarReplacer = strings.NewReplacer("&{", "${")

func dataSourceScalewayObjectPolicyDocument() *schema.Resource {
	setOfString := &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	}

	return &schema.Resource{
		ReadContext: dataSourcePolicyDocumentRead,

		Schema: map[string]*schema.Schema{
			"json": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"override_policy_documents": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"policy_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"source_policy_documents": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"statement": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"actions": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
								ValidateFunc: validation.StringInSlice([]string{
									"*",
									"s3:*",
									"s3:AbortMultipartUpload",
									"s3:DeleteBucketWebsite",
									"s3:DeleteObject",
									"s3:DeleteObjectTagging",
									"s3:DeleteObjectVersion",
									"s3:DeleteObjectVersionTagging",
									"s3:GetBucketTagging",
									"s3:GetBucketVersioning",
									"s3:GetBucketWebsite",
									"s3:GetLifecycleConfiguration",
									"s3:GetObject",
									"s3:GetObjectTagging",
									"s3:GetObjectVersion",
									"s3:GetObjectVersionTagging",
									"s3:ListBucket",
									"s3:ListBucketMultipartUploads",
									"s3:ListMultipartUploadParts",
									"s3:PutBucketTagging",
									"s3:PutBucketVersioning",
									"s3:PutBucketWebsite",
									"s3:PutLifecycleConfiguration",
									"s3:PutObject",
									"s3:PutObjectTagging",
									"s3:PutObjectVersionTagging",
								}, false),
							},
						},
						"condition": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"test": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.StringInSlice([]string{
											"DateGreaterThan",
											"DateGreaterThanEquals",
											"DateLessThan",
											"DateLessThanEquals",
											"IpAddress",
											"NotIpAddress",
											"StringEquals",
											"StringEqualsIgnoreCase",
											"StringLike",
											"StringNotEquals",
											"StringNotEqualsIgnoreCase",
											"StringNotLike",
										}, false),
									},
									"values": {
										Type:     schema.TypeList,
										Required: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"variable": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.StringInSlice([]string{
											"aws:SourceIp",
											"aws:Referer",
											"aws:CurrentTime",
											"aws:EpochTime",
										}, false),
									},
								},
							},
						},
						"effect": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "Allow",
							ValidateFunc: validation.StringInSlice([]string{"Allow", "Deny"}, false),
						},
						"not_actions":    setOfString,
						"not_principals": dataSourcePolicyPrincipalSchema(),
						"not_resources":  setOfString,
						"principals":     dataSourcePolicyPrincipalSchema(),
						"resources":      setOfString,
						"sid": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"version": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "2012-10-17",
				ValidateFunc: validation.StringInSlice([]string{
					"2012-10-17",
				}, false),
			},
		},
	}
}

func dataSourcePolicyDocumentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	mergedDoc := &IAMPolicyDoc{}

	if v, ok := d.GetOk("source_json"); ok {
		if err := json.Unmarshal([]byte(v.(string)), mergedDoc); err != nil {
			return diag.FromErr(err)
		}
	}

	if v, ok := d.GetOk("source_policy_documents"); ok && len(v.([]interface{})) > 0 {
		// generate sid map to assure there are no duplicates in source jsons
		sidMap := make(map[string]struct{})
		for _, stmt := range mergedDoc.Statements {
			if stmt.Sid != "" {
				sidMap[stmt.Sid] = struct{}{}
			}
		}

		// merge sourceDocs in order specified
		for sourceJSONIndex, sourceJSON := range v.([]interface{}) {
			sourceDoc := &IAMPolicyDoc{}
			if err := json.Unmarshal([]byte(sourceJSON.(string)), sourceDoc); err != nil {
				return diag.FromErr(err)
			}

			// assure all statements in sourceDoc are unique before merging
			for stmtIndex, stmt := range sourceDoc.Statements {
				if stmt.Sid != "" {
					if _, sidExists := sidMap[stmt.Sid]; sidExists {
						return diag.FromErr(fmt.Errorf("duplicate Sid (%s) in source_policy_documents (item %d; statement %d). Remove the Sid or ensure Sids are unique", stmt.Sid, sourceJSONIndex, stmtIndex))
					}
					sidMap[stmt.Sid] = struct{}{}
				}
			}

			mergedDoc.Merge(sourceDoc)
		}
	}

	// process the current document
	doc := &IAMPolicyDoc{
		Version: d.Get("version").(string),
	}

	if policyID, hasPolicyID := d.GetOk("policy_id"); hasPolicyID {
		doc.Id = policyID.(string)
	}

	if cfgStmts, hasCfgStmts := d.GetOk("statement"); hasCfgStmts {
		cfgStmtIntf := cfgStmts.([]interface{})
		stmts := make([]*IAMPolicyStatement, len(cfgStmtIntf))
		sidMap := make(map[string]struct{})

		for i, stmtI := range cfgStmtIntf {
			cfgStmt := stmtI.(map[string]interface{})
			stmt := &IAMPolicyStatement{
				Effect: cfgStmt["effect"].(string),
			}

			if sid, ok := cfgStmt["sid"]; ok {
				if _, ok := sidMap[sid.(string)]; ok {
					return diag.FromErr(fmt.Errorf("duplicate Sid (%s). Remove the Sid or ensure the Sid is unique", sid.(string)))
				}
				stmt.Sid = sid.(string)
				if len(stmt.Sid) > 0 {
					sidMap[stmt.Sid] = struct{}{}
				}
			}

			if actions := cfgStmt["actions"].(*schema.Set).List(); len(actions) > 0 {
				stmt.Actions = policyDecodeConfigStringList(actions)
			}
			if actions := cfgStmt["not_actions"].(*schema.Set).List(); len(actions) > 0 {
				stmt.NotActions = policyDecodeConfigStringList(actions)
			}

			if resources := cfgStmt["resources"].(*schema.Set).List(); len(resources) > 0 {
				var err error
				stmt.Resources, err = dataSourcePolicyDocumentReplaceVarsInList(
					policyDecodeConfigStringList(resources), doc.Version,
				)
				if err != nil {
					return diag.FromErr(fmt.Errorf("error reading resources: %w", err))
				}
			}
			if notResources := cfgStmt["not_resources"].(*schema.Set).List(); len(notResources) > 0 {
				var err error
				stmt.NotResources, err = dataSourcePolicyDocumentReplaceVarsInList(
					policyDecodeConfigStringList(notResources), doc.Version,
				)
				if err != nil {
					return diag.FromErr(fmt.Errorf("error reading not_resources: %w", err))
				}
			}

			if principals := cfgStmt["principals"].(*schema.Set).List(); len(principals) > 0 {
				var err error
				stmt.Principals, err = dataSourcePolicyDocumentMakePrincipals(principals, doc.Version)
				if err != nil {
					return diag.FromErr(fmt.Errorf("error reading principals: %w", err))
				}
			}

			if notPrincipals := cfgStmt["not_principals"].(*schema.Set).List(); len(notPrincipals) > 0 {
				var err error
				stmt.NotPrincipals, err = dataSourcePolicyDocumentMakePrincipals(notPrincipals, doc.Version)
				if err != nil {
					return diag.FromErr(fmt.Errorf("error reading not_principals: %w", err))
				}
			}

			if conditions := cfgStmt["condition"].(*schema.Set).List(); len(conditions) > 0 {
				var err error
				stmt.Conditions, err = dataSourcePolicyDocumentMakeConditions(conditions, doc.Version)
				if err != nil {
					return diag.FromErr(fmt.Errorf("error reading condition: %w", err))
				}
			}

			stmts[i] = stmt
		}

		doc.Statements = stmts
	}

	// merge our current document into mergedDoc
	mergedDoc.Merge(doc)

	// merge override_policy_documents policies into mergedDoc in order specified
	if v, ok := d.GetOk("override_policy_documents"); ok && len(v.([]interface{})) > 0 {
		for _, overrideJSON := range v.([]interface{}) {
			overrideDoc := &IAMPolicyDoc{}
			if err := json.Unmarshal([]byte(overrideJSON.(string)), overrideDoc); err != nil {
				return diag.FromErr(err)
			}

			mergedDoc.Merge(overrideDoc)
		}
	}

	// merge in override_json
	if v, ok := d.GetOk("override_json"); ok {
		overrideDoc := &IAMPolicyDoc{}
		if err := json.Unmarshal([]byte(v.(string)), overrideDoc); err != nil {
			return diag.FromErr(err)
		}

		mergedDoc.Merge(overrideDoc)
	}

	jsonDoc, err := json.MarshalIndent(mergedDoc, "", "  ")
	if err != nil {
		// should never happen if the above code is correct
		return diag.FromErr(err)
	}
	jsonString := string(jsonDoc)

	_ = d.Set("json", jsonString)
	d.SetId(strconv.Itoa(StringHashcode(jsonString)))

	return nil
}

func dataSourcePolicyDocumentReplaceVarsInList(in interface{}, version string) (interface{}, error) {
	switch v := in.(type) {
	case string:
		if version == "2008-10-17" && strings.Contains(v, "&{") {
			return nil, fmt.Errorf("found &{ sequence in (%s), which is not supported in document version 2008-10-17", v)
		}
		return dataSourcePolicyDocumentVarReplacer.Replace(v), nil
	case []string:
		out := make([]string, len(v))
		for i, item := range v {
			if version == "2008-10-17" && strings.Contains(item, "&{") {
				return nil, fmt.Errorf("found &{ sequence in (%s), which is not supported in document version 2008-10-17", item)
			}
			out[i] = dataSourcePolicyDocumentVarReplacer.Replace(item)
		}
		return out, nil
	default:
		return nil, errors.New("dataSourcePolicyDocumentReplaceVarsInList: input not string nor []string")
	}
}

func dataSourcePolicyDocumentMakeConditions(in []interface{}, version string) (IAMPolicyStatementConditionSet, error) {
	out := make([]IAMPolicyStatementCondition, len(in))
	for i, itemI := range in {
		var err error
		item := itemI.(map[string]interface{})
		out[i] = IAMPolicyStatementCondition{
			Test:     item["test"].(string),
			Variable: item["variable"].(string),
		}
		out[i].Values, err = dataSourcePolicyDocumentReplaceVarsInList(
			aws.StringValueSlice(expandStringListKeepEmpty(item["values"].([]interface{}))),
			version,
		)
		if err != nil {
			return nil, fmt.Errorf("error reading values: %w", err)
		}
		itemValues := out[i].Values.([]string)
		if len(itemValues) == 1 {
			out[i].Values = itemValues[0]
		}
	}
	return out, nil
}

func dataSourcePolicyDocumentMakePrincipals(in []interface{}, version string) (IAMPolicyStatementPrincipalSet, error) {
	out := make([]IAMPolicyStatementPrincipal, len(in))
	for i, itemI := range in {
		var err error
		item := itemI.(map[string]interface{})
		out[i] = IAMPolicyStatementPrincipal{
			Type: item["type"].(string),
		}
		out[i].Identifiers, err = dataSourcePolicyDocumentReplaceVarsInList(
			policyDecodeConfigStringList(
				item["identifiers"].(*schema.Set).List(),
			), version,
		)
		if err != nil {
			return nil, fmt.Errorf("error reading identifiers: %w", err)
		}
	}
	return out, nil
}

func dataSourcePolicyPrincipalSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"type": {
					Type:     schema.TypeString,
					Required: true,
				},
				"identifiers": {
					Type:     schema.TypeSet,
					Required: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
			},
		},
	}
}

type IAMPolicyDoc struct {
	Version    string                `json:",omitempty"`
	Id         string                `json:",omitempty"` //nolint:golint,stylecheck
	Statements []*IAMPolicyStatement `json:"Statement"`
}

func (s *IAMPolicyDoc) Merge(newDoc *IAMPolicyDoc) {
	// adopt newDoc's Id
	if len(newDoc.Id) > 0 {
		s.Id = newDoc.Id
	}

	// let newDoc upgrade our Version
	if newDoc.Version > s.Version {
		s.Version = newDoc.Version
	}

	// merge in newDoc's statements, overwriting any existing Sids
	var seen bool
	for _, newStatement := range newDoc.Statements {
		if len(newStatement.Sid) == 0 {
			s.Statements = append(s.Statements, newStatement)
			continue
		}
		seen = false
		for i, existingStatement := range s.Statements {
			if existingStatement.Sid == newStatement.Sid {
				s.Statements[i] = newStatement
				seen = true
				break
			}
		}
		if !seen {
			s.Statements = append(s.Statements, newStatement)
		}
	}
}

type IAMPolicyStatement struct {
	Sid           string
	Effect        string                         `json:",omitempty"`
	Actions       interface{}                    `json:"Action,omitempty"`
	NotActions    interface{}                    `json:"NotAction,omitempty"`
	Resources     interface{}                    `json:"Resource,omitempty"`
	NotResources  interface{}                    `json:"NotResource,omitempty"`
	Principals    IAMPolicyStatementPrincipalSet `json:"Principal,omitempty"`
	NotPrincipals IAMPolicyStatementPrincipalSet `json:"NotPrincipal,omitempty"`
	Conditions    IAMPolicyStatementConditionSet `json:"condition,omitempty"`
}

type (
	IAMPolicyStatementPrincipalSet []IAMPolicyStatementPrincipal
	IAMPolicyStatementConditionSet []IAMPolicyStatementCondition
)

type IAMPolicyStatementPrincipal struct {
	Type        string
	Identifiers interface{}
}

type IAMPolicyStatementCondition struct {
	Test     string
	Variable string
	Values   interface{}
}

func policyDecodeConfigStringList(lI []interface{}) interface{} {
	if len(lI) == 1 {
		return lI[0].(string)
	}
	ret := make([]string, len(lI))
	for i, vI := range lI {
		ret[i] = vI.(string)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(ret)))
	return ret
}

func expandStringListKeepEmpty(configured []interface{}) []*string {
	vs := make([]*string, 0, len(configured))
	for _, v := range configured {
		if val, ok := v.(string); ok {
			vs = append(vs, aws.String(val))
		}
	}
	return vs
}
