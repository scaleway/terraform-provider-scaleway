package scaleway

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	BucketACLSeparator = "/"
)

func resourceScalewayObjectBucketACL() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBucketACLCreate,
		ReadContext:   resourceBucketACLRead,
		UpdateContext: resourceBucketACLUpdate,
		DeleteContext: resourceBucketACLDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"access_control_policy": {
				Type:          schema.TypeList,
				Optional:      true,
				Computed:      true,
				MaxItems:      1,
				ConflictsWith: []string{"acl"},
				Description:   "A configuration block that sets the ACL permissions for an object per grantee.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"grant": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"grantee": {
										Type:        schema.TypeList,
										Optional:    true,
										MaxItems:    1,
										Description: "Configuration block for the project being granted permissions.",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"display_name": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"id": {
													Type:         schema.TypeString,
													Required:     true,
													Description:  "The project ID owner of the grantee.",
													ValidateFunc: validationUUID(),
												},
												"type": {
													Type:         schema.TypeString,
													Required:     true,
													Description:  "Type of grantee. Valid values: `CanonicalUser`",
													ValidateFunc: validation.StringInSlice([]string{s3.TypeCanonicalUser}, false),
												},
											},
										},
									},
									"permission": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(s3.Permission_Values(), false),
										Description:  "Logging permissions assigned to the grantee for the bucket.",
									},
								},
							},
						},
						"owner": {
							Type:        schema.TypeList,
							Required:    true,
							MaxItems:    1,
							Description: "Configuration block of the bucket project owner's display organization ID.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"display_name": {
										Type:         schema.TypeString,
										Computed:     true,
										Optional:     true,
										Description:  "The project ID of the grantee.",
										ValidateFunc: validationUUID(),
									},
									"id": {
										Type:         schema.TypeString,
										Required:     true,
										Description:  "The display ID of the project.",
										ValidateFunc: validationUUID(),
									},
								},
							},
						},
					},
				},
			},
			"acl": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ACL of the bucket: either 'public-read' or 'private'.",
				ValidateFunc: validation.StringInSlice([]string{
					s3.ObjectCannedACLPrivate,
					s3.ObjectCannedACLPublicRead,
					s3.ObjectCannedACLPublicReadWrite,
					s3.ObjectCannedACLAuthenticatedRead,
				}, false),
			},
			"bucket": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 63),
				Description:  "The bucket name.",
			},
			"expected_bucket_owner": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Description:  "The project ID as owner.",
				ValidateFunc: validationUUID(),
			},
			"region": regionSchema(),
		},
	}
}

func resourceBucketACLCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn, region, err := s3ClientWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	bucket := expandID(d.Get("bucket").(string))
	expectedBucketOwner := d.Get("expected_bucket_owner").(string)
	acl := d.Get("acl").(string)

	input := &s3.PutBucketAclInput{
		Bucket: aws.String(bucket),
	}

	if acl != "" {
		input.ACL = aws.String(acl)
	}

	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	if v, ok := d.GetOk("access_control_policy"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.AccessControlPolicy = expandBucketACLAccessControlPolicy(v.([]interface{}))
	}

	out, err := retryOnAWSCode(ctx, s3.ErrCodeNoSuchBucket, func() (interface{}, error) {
		out, err := conn.PutBucketAclWithContext(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("error occurred while doing PutBucketAclWithContext: %w", err)
		}
		return out, nil
	})
	if err != nil {
		return diag.FromErr(fmt.Errorf("error putting Object Storage ACL: %s", err))
	}
	tflog.Debug(ctx, fmt.Sprintf("output: %v", out))

	d.SetId(BucketACLCreateResourceID(region, bucket, acl))

	return resourceBucketACLRead(ctx, d, meta)
}

func expandBucketACLAccessControlPolicy(l []interface{}) *s3.AccessControlPolicy {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &s3.AccessControlPolicy{}

	if v, ok := tfMap["grant"].(*schema.Set); ok && v.Len() > 0 {
		result.Grants = expandBucketACLAccessControlPolicyGrants(v.List())
	}

	if v, ok := tfMap["owner"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.Owner = expandBucketACLAccessControlPolicyOwner(v)
	}

	return result
}

func expandBucketACLAccessControlPolicyGrants(l []interface{}) []*s3.Grant {
	var grants []*s3.Grant

	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		grant := &s3.Grant{}

		if v, ok := tfMap["grantee"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			grant.Grantee = expandBucketACLAccessControlPolicyGrantsGrantee(v)
		}

		if v, ok := tfMap["permission"].(string); ok && v != "" {
			grant.Permission = aws.String(v)
		}

		grants = append(grants, grant)
	}

	return grants
}

func expandBucketACLAccessControlPolicyGrantsGrantee(l []interface{}) *s3.Grantee {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &s3.Grantee{}

	if v, ok := tfMap["id"].(string); ok && v != "" {
		result.ID = buildBucketOwnerID(aws.String(v))
	}

	if v, ok := tfMap["type"].(string); ok && v != "" {
		result.Type = aws.String(v)
	}

	return result
}

func expandBucketACLAccessControlPolicyOwner(l []interface{}) *s3.Owner {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	owner := &s3.Owner{}

	if v, ok := tfMap["display_name"].(string); ok && v != "" {
		owner.DisplayName = buildBucketOwnerID(aws.String(v))
	}

	if v, ok := tfMap["id"].(string); ok && v != "" {
		owner.ID = buildBucketOwnerID(aws.String(v))
	}

	return owner
}

func flattenBucketACLAccessControlPolicy(output *s3.GetBucketAclOutput) []interface{} {
	if output == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if len(output.Grants) > 0 {
		m["grant"] = flattenBucketACLAccessControlPolicyGrants(output.Grants)
	}

	if output.Owner != nil {
		m["owner"] = flattenBucketACLAccessControlPolicyOwner(output.Owner)
	}

	return []interface{}{m}
}

func flattenBucketACLAccessControlPolicyGrants(grants []*s3.Grant) []interface{} {
	var results []interface{}

	for _, grant := range grants {
		if grant == nil {
			continue
		}

		m := make(map[string]interface{})

		if grant.Grantee != nil {
			m["grantee"] = flattenBucketACLAccessControlPolicyGrantsGrantee(grant.Grantee)
		}

		if grant.Permission != nil {
			m["permission"] = aws.StringValue(grant.Permission)
		}

		results = append(results, m)
	}

	return results
}

func flattenBucketACLAccessControlPolicyGrantsGrantee(grantee *s3.Grantee) []interface{} {
	if grantee == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if grantee.DisplayName != nil {
		m["display_name"] = aws.StringValue(normalizeOwnerID(grantee.DisplayName))
	}

	if grantee.ID != nil {
		m["id"] = aws.StringValue(normalizeOwnerID(grantee.ID))
	}

	if grantee.Type != nil {
		m["type"] = aws.StringValue(grantee.Type)
	}

	return []interface{}{m}
}

func flattenBucketACLAccessControlPolicyOwner(owner *s3.Owner) []interface{} {
	if owner == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if owner.DisplayName != nil {
		m["display_name"] = aws.StringValue(normalizeOwnerID(owner.DisplayName))
	}

	if owner.ID != nil {
		m["id"] = aws.StringValue(normalizeOwnerID(owner.ID))
	}

	return []interface{}{m}
}

func resourceBucketACLRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	expectedBucketOwner := d.Get("expected_bucket_owner")
	conn, region, bucket, acl, err := s3ClientWithRegionWithNameACL(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	input := &s3.GetBucketAclInput{
		Bucket: aws.String(bucket),
	}

	if v, ok := d.GetOk("expected_bucket_owner"); ok {
		input.ExpectedBucketOwner = aws.String(v.(string))
	}

	output, err := conn.GetBucketAclWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		tflog.Warn(ctx, fmt.Sprintf("[WARN] Object Bucket ACL (%s) not found, removing from state", d.Id()))
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting object storage bucket ACL (%s): %w", d.Id(), err))
	}

	if output == nil {
		return diag.FromErr(fmt.Errorf("error getting object bucket ACL (%s): empty output", d.Id()))
	}

	_ = d.Set("acl", acl)
	_ = d.Set("expected_bucket_owner", expectedBucketOwner)
	if err := d.Set("access_control_policy", flattenBucketACLAccessControlPolicy(output)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting access_control_policy: %w", err))
	}
	_ = d.Set("region", region)
	_ = d.Set("bucket", expandID(bucket))

	return nil
}

// BucketACLCreateResourceID is a method for creating an ID string
// with the bucket name and optional organizationID and/or ACL.
func BucketACLCreateResourceID(region scw.Region, bucket, acl string) string {
	if acl == "" {
		return newRegionalIDString(region, bucket)
	}
	return newRegionalIDString(region, strings.Join([]string{bucket, acl}, BucketACLSeparator))
}

func resourceBucketACLUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn, region, bucket, acl, err := s3ClientWithRegionWithNameACL(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	input := &s3.PutBucketAclInput{
		Bucket: aws.String(bucket),
	}

	if d.HasChange("acl") {
		acl = d.Get("acl").(string)
		input.ACL = aws.String(acl)
	}

	if ok := d.HasChange("expected_bucket_owner"); ok {
		input.ExpectedBucketOwner = aws.String(d.Get("expected_bucket_owner").(string))
	}

	if d.HasChange("access_control_policy") {
		input.AccessControlPolicy = expandBucketACLAccessControlPolicy(d.Get("access_control_policy").([]interface{}))
	}

	_, err = conn.PutBucketAclWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating object bucket ACL (%s): %w", d.Id(), err))
	}

	if d.HasChange("acl") {
		// Set new ACL value back in resource ID
		d.SetId(BucketACLCreateResourceID(region, bucket, acl))
	}

	return resourceBucketACLRead(ctx, d, meta)
}

func resourceBucketACLDelete(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
	tflog.Warn(ctx, "[WARN] Cannot destroy Object Bucket ACL. Terraform will remove this resource from the state file, however resources may remain.")
	return nil
}
