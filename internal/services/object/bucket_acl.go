package object

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

const (
	BucketACLSeparator = "/"

	// AWS predefined group URIs for bucket ACL grants
	AuthenticatedUsersURI = "http://acs.amazonaws.com/groups/global/AuthenticatedUsers"
	AllUsersURI           = "http://acs.amazonaws.com/groups/global/AllUsers"
)

func ResourceBucketACL() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBucketACLCreate,
		ReadContext:   resourceBucketACLRead,
		UpdateContext: resourceBucketACLUpdate,
		DeleteContext: resourceBucketACLDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaFunc: bucketAclSchema,
	}
}

func bucketAclSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
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
						Type:        schema.TypeSet,
						Description: "Grant",
						Optional:    true,
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
												Type:        schema.TypeString,
												Description: "Display name of the grantee to grant access to.",
												Computed:    true,
											},
											"uri": {
												Type:        schema.TypeString,
												Optional:    true,
												Description: "The uri of the grantee if you are granting permissions to a predefined group.",
												ValidateFunc: validation.StringInSlice([]string{
													AuthenticatedUsersURI,
													AllUsersURI,
												}, false),
											},
											"id": {
												Type:             schema.TypeString,
												Optional:         true,
												Description:      "The project ID owner of the grantee.",
												ValidateDiagFunc: verify.IsUUID(),
											},
											"type": {
												Type:         schema.TypeString,
												Optional:     true,
												Description:  "Type of grantee. Valid values: `CanonicalUser`, `Group`",
												ValidateFunc: validation.StringInSlice([]string{string(s3Types.TypeCanonicalUser), string(s3Types.TypeGroup)}, false),
											},
										},
									},
								},
								"permission": {
									Type:     schema.TypeString,
									Required: true,
									ValidateFunc: validation.StringInSlice([]string{
										string(s3Types.PermissionFullControl),
										string(s3Types.PermissionRead),
										string(s3Types.PermissionWrite),
										string(s3Types.PermissionReadAcp),
										string(s3Types.PermissionWriteAcp),
									}, false),
									Description: "Logging permissions assigned to the grantee for the bucket.",
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
									Type:             schema.TypeString,
									Computed:         true,
									Optional:         true,
									Description:      "The project ID of the grantee.",
									ValidateDiagFunc: verify.IsUUID(),
								},
								"id": {
									Type:             schema.TypeString,
									Required:         true,
									Description:      "The display ID of the project.",
									ValidateDiagFunc: verify.IsUUID(),
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
			Description: "ACL of the bucket: either 'private', 'public-read', 'public-read-write' or 'authenticated-read'.",
			ValidateFunc: validation.StringInSlice([]string{
				string(s3Types.ObjectCannedACLPrivate),
				string(s3Types.ObjectCannedACLPublicRead),
				string(s3Types.ObjectCannedACLPublicReadWrite),
				string(s3Types.ObjectCannedACLAuthenticatedRead),
			}, false),
		},
		"bucket": {
			Type:             schema.TypeString,
			Required:         true,
			ForceNew:         true,
			ValidateFunc:     validation.StringLenBetween(1, 63),
			Description:      "The bucket's name or regional ID.",
			DiffSuppressFunc: dsf.Locality,
		},
		"expected_bucket_owner": {
			Type:             schema.TypeString,
			Optional:         true,
			ForceNew:         true,
			Description:      "The project ID as owner.",
			ValidateDiagFunc: verify.IsUUID(),
		},
		"region":     regional.Schema(),
		"project_id": account.ProjectIDSchema(),
	}
}

func resourceBucketACLCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	conn, region, err := s3ClientWithRegion(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	regionalID := regional.ExpandID(d.Get("bucket"))
	bucket := regionalID.ID
	bucketRegion := regionalID.Region

	if bucketRegion != "" && bucketRegion != region {
		conn, err = s3ClientForceRegion(ctx, d, m, bucketRegion.String())
		if err != nil {
			return diag.FromErr(err)
		}

		region = bucketRegion
	}

	expectedBucketOwner := d.Get("expected_bucket_owner").(string)
	acl := d.Get("acl").(string)

	input := &s3.PutBucketAclInput{
		Bucket: aws.String(bucket),
	}

	if acl != "" {
		input.ACL = s3Types.BucketCannedACL(acl)
	}

	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	if v, ok := d.GetOk("access_control_policy"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		accessControlPolicy, err := expandAndValidateBucketACLAccessControlPolicy(v.([]any))
		if err != nil {
			return diag.FromErr(err)
		}

		input.AccessControlPolicy = accessControlPolicy
	}

	out, err := conn.PutBucketAcl(ctx, input)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error putting Object Storage ACL: %w", err))
	}

	tflog.Debug(ctx, fmt.Sprintf("output: %v", out))

	d.SetId(BucketACLCreateResourceID(region, bucket, acl))

	return resourceBucketACLRead(ctx, d, m)
}

func expandAndValidateBucketACLAccessControlPolicy(l []any) (*s3Types.AccessControlPolicy, error) {
	if len(l) == 0 || l[0] == nil {
		return nil, nil
	}

	tfMap, ok := l[0].(map[string]any)
	if !ok {
		return nil, nil
	}

	result := &s3Types.AccessControlPolicy{}

	if v, ok := tfMap["grant"].(*schema.Set); ok && v.Len() > 0 {
		grants, err := expandAndValidateBucketACLAccessControlPolicyGrants(v.List())
		if err != nil {
			return nil, err
		}

		result.Grants = grants
	}

	if v, ok := tfMap["owner"].([]any); ok && len(v) > 0 && v[0] != nil {
		result.Owner = expandBucketACLAccessControlPolicyOwner(v)
	}

	return result, nil
}

func expandAndValidateBucketACLAccessControlPolicyGrants(l []any) ([]s3Types.Grant, error) {
	grants := make([]s3Types.Grant, 0, len(l))

	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		grant := s3Types.Grant{}

		if v, ok := tfMap["grantee"].([]any); ok && len(v) > 0 && v[0] != nil {
			grantee, err := expandAndValidateBucketACLAccessControlPolicyGrantsGrantee(v)
			if err != nil {
				return nil, err
			}

			grant.Grantee = grantee
		}

		if v, ok := tfMap["permission"].(string); ok && v != "" {
			grant.Permission = s3Types.Permission(v)
		}

		grants = append(grants, grant)
	}

	return grants, nil
}

func expandAndValidateBucketACLAccessControlPolicyGrantsGrantee(l []any) (*s3Types.Grantee, error) {
	if len(l) == 0 || l[0] == nil {
		return nil, nil
	}

	tfMap, ok := l[0].(map[string]any)
	if !ok {
		return nil, nil
	}

	result := &s3Types.Grantee{}

	if v, ok := tfMap["id"].(string); ok && v != "" {
		result.ID = buildBucketOwnerID(aws.String(v))
	}

	if v, ok := tfMap["uri"].(string); ok && v != "" {
		result.URI = aws.String(v)
	}

	if v, ok := tfMap["type"].(string); ok && v != "" {
		result.Type = s3Types.Type(v)
	}

	if result.Type == s3Types.TypeCanonicalUser && result.ID == nil {
		return nil, errors.New("id is required when grantee type is CanonicalUser")
	}

	if result.Type == s3Types.TypeGroup && result.URI == nil {
		return nil, errors.New("uri is required when grantee type is Group")
	}

	return result, nil
}

func expandBucketACLAccessControlPolicyOwner(l []any) *s3Types.Owner {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]any)
	if !ok {
		return nil
	}

	owner := &s3Types.Owner{}

	if v, ok := tfMap["display_name"].(string); ok && v != "" {
		owner.DisplayName = buildBucketOwnerID(aws.String(v))
	}

	if v, ok := tfMap["id"].(string); ok && v != "" {
		owner.ID = buildBucketOwnerID(aws.String(v))
	}

	return owner
}

func flattenBucketACLAccessControlPolicy(output *s3.GetBucketAclOutput) []any {
	if output == nil {
		return []any{}
	}

	m := make(map[string]any)

	if len(output.Grants) > 0 {
		m["grant"] = flattenBucketACLAccessControlPolicyGrants(output.Grants)
	}

	if output.Owner != nil {
		m["owner"] = flattenBucketACLAccessControlPolicyOwner(output.Owner)
	}

	return []any{m}
}

func flattenBucketACLAccessControlPolicyGrants(grants []s3Types.Grant) []any {
	results := make([]any, 0, len(grants))

	for _, grant := range grants {
		if grant.Grantee == nil && grant.Permission == "" {
			continue
		}

		m := make(map[string]any)

		if grant.Grantee != nil {
			m["grantee"] = flattenBucketACLAccessControlPolicyGrantsGrantee(grant.Grantee)
		}

		if grant.Permission != "" {
			m["permission"] = grant.Permission
		}

		results = append(results, m)
	}

	return results
}

func flattenBucketACLAccessControlPolicyGrantsGrantee(grantee *s3Types.Grantee) []any {
	if grantee == nil {
		return []any{}
	}

	m := make(map[string]any)

	if grantee.DisplayName != nil {
		m["display_name"] = NormalizeOwnerID(grantee.DisplayName)
	}

	if grantee.URI != nil {
		m["uri"] = *grantee.URI
	}

	if grantee.ID != nil {
		m["id"] = NormalizeOwnerID(grantee.ID)
	}

	if grantee.Type != "" {
		m["type"] = grantee.Type
	}

	return []any{m}
}

func flattenBucketACLAccessControlPolicyOwner(owner *s3Types.Owner) []any {
	if owner == nil {
		return []any{}
	}

	m := make(map[string]any)

	if owner.DisplayName != nil {
		m["display_name"] = NormalizeOwnerID(owner.DisplayName)
	}

	if owner.ID != nil {
		m["id"] = NormalizeOwnerID(owner.ID)
	}

	return []any{m}
}

func resourceBucketACLRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	expectedBucketOwner := d.Get("expected_bucket_owner")

	conn, region, bucket, acl, err := s3ClientWithRegionWithNameACL(ctx, d, m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	input := &s3.GetBucketAclInput{
		Bucket: aws.String(bucket),
	}

	if v, ok := d.GetOk("expected_bucket_owner"); ok {
		input.ExpectedBucketOwner = aws.String(v.(string))
	}

	output, err := conn.GetBucketAcl(ctx, input)

	if !d.IsNewResource() && errors.As(err, new(*s3Types.NoSuchBucket)) {
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
	_ = d.Set("project_id", NormalizeOwnerID(output.Owner.ID))
	_ = d.Set("bucket", locality.ExpandID(bucket))

	return nil
}

// BucketACLCreateResourceID is a method for creating an ID string
// with the bucket name and optional organizationID and/or ACL.
func BucketACLCreateResourceID(region scw.Region, bucket, acl string) string {
	if acl == "" {
		return regional.NewIDString(region, bucket)
	}

	return regional.NewIDString(region, strings.Join([]string{bucket, acl}, BucketACLSeparator))
}

func resourceBucketACLUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	conn, region, bucket, acl, err := s3ClientWithRegionWithNameACL(ctx, d, m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	input := &s3.PutBucketAclInput{
		Bucket: aws.String(bucket),
	}

	if d.HasChange("acl") {
		acl = d.Get("acl").(string)
		input.ACL = s3Types.BucketCannedACL(acl)
	}

	if ok := d.HasChange("expected_bucket_owner"); ok {
		input.ExpectedBucketOwner = aws.String(d.Get("expected_bucket_owner").(string))
	}

	if d.HasChange("access_control_policy") {
		accessControlPolicy, err := expandAndValidateBucketACLAccessControlPolicy(d.Get("access_control_policy").([]any))
		if err != nil {
			return diag.FromErr(err)
		}

		input.AccessControlPolicy = accessControlPolicy
	}

	_, err = conn.PutBucketAcl(ctx, input)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating object bucket ACL (%s): %w", d.Id(), err))
	}

	if d.HasChange("acl") {
		// Set new ACL value back in resource ID
		d.SetId(BucketACLCreateResourceID(region, bucket, acl))
	}

	return resourceBucketACLRead(ctx, d, m)
}

func resourceBucketACLDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	conn, _, bucket, _, err := s3ClientWithRegionWithNameACL(ctx, d, m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = conn.PutBucketAcl(ctx, &s3.PutBucketAclInput{
		Bucket: &bucket,
		ACL:    s3Types.BucketCannedACL(s3Types.ObjectCannedACLPrivate),
	})
	if err != nil {
		return diag.FromErr(fmt.Errorf("error putting bucket ACL: %w", err))
	}

	return diag.Diagnostics{
		diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Deleting Object Bucket ACL resource resets ACL to private",
			Detail:   "Deleting Object Bucket ACL resource resets the bucket's ACL to its default value: private.\nIf you wish to set it to something else, you should recreate a Bucket ACL resource with the `acl` field filled accordingly.",
		},
	}
}
