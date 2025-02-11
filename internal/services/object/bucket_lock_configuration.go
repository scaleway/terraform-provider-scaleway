package object

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
)

func ResourceLockConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceObjectLockConfigurationCreate,
		ReadContext:   resourceObjectLockConfigurationRead,
		UpdateContext: resourceObjectLockConfigurationUpdate,
		DeleteContext: resourceObjectLockConfigurationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateFunc:     validation.StringLenBetween(1, 63),
				Description:      "The bucket's name or regional ID.",
				DiffSuppressFunc: dsf.Locality,
			},
			"rule": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"default_retention": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"mode": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice([]string{"GOVERNANCE", "COMPLIANCE"}, false),
										Description:  "The default Object Lock retention mode you want to apply to new objects placed in the specified bucket.",
									},
									"days": {
										Type:          schema.TypeInt,
										Optional:      true,
										Description:   "The number of days that you want to specify for the default retention period.",
										ConflictsWith: []string{"rule.0.default_retention.0.years"},
									},
									"years": {
										Type:          schema.TypeInt,
										Optional:      true,
										Description:   "The number of years that you want to specify for the default retention period.",
										ConflictsWith: []string{"rule.0.default_retention.0.days"},
									},
								},
							},
						},
					},
				},
				Description: "Specifies the Object Lock rule for the specified object.",
			},
			"region":     regional.Schema(),
			"project_id": account.ProjectIDSchema(),
		},
	}
}

func resourceObjectLockConfigurationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

	input := &s3.PutObjectLockConfigurationInput{
		Bucket: aws.String(bucket),
		ObjectLockConfiguration: &s3Types.ObjectLockConfiguration{
			ObjectLockEnabled: "Enabled",
			Rule:              expandBucketLockConfigurationRule(d.Get("rule").([]interface{})),
		},
	}

	_, err = conn.PutObjectLockConfiguration(ctx, input)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating object bucket (%s) lock configuration: %w", bucket, err))
	}

	d.SetId(regional.NewIDString(region, bucket))

	return resourceObjectLockConfigurationRead(ctx, d, m)
}

func resourceObjectLockConfigurationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	conn, _, bucket, err := s3ClientWithRegionAndName(ctx, d, m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	input := &s3.GetObjectLockConfigurationInput{
		Bucket: aws.String(bucket),
	}

	output, err := conn.GetObjectLockConfiguration(ctx, input)
	if !d.IsNewResource() && errors.As(err, new(*s3Types.NoSuchBucket)) {
		tflog.Warn(ctx, fmt.Sprintf("Object Bucket Lock Configuration (%s) not found, removing from state", d.Id()))
		d.SetId("")

		return nil
	}

	if output == nil {
		if d.IsNewResource() {
			return diag.FromErr(fmt.Errorf("error reading object bucket lock configuration (%s): empty output", d.Id()))
		}

		tflog.Warn(ctx, fmt.Sprintf("Object Bucket Lock Configuration (%s) not found, removing from state", d.Id()))
		d.SetId("")

		return nil
	}

	acl, err := conn.GetBucketAcl(ctx, &s3.GetBucketAclInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		return diag.FromErr(fmt.Errorf("couldn't read bucket acl: %s", err))
	}

	_ = d.Set("project_id", NormalizeOwnerID(acl.Owner.ID))

	_ = d.Set("bucket", bucket)
	if output.ObjectLockConfiguration != nil {
		_ = d.Set("rule", flattenBucketLockConfigurationRule(output.ObjectLockConfiguration.Rule))
	} else {
		_ = d.Set("rule", flattenBucketLockConfigurationRule(nil))
	}

	return nil
}

func resourceObjectLockConfigurationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	conn, _, bucket, err := s3ClientWithRegionAndName(ctx, d, m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	lockConfig := &s3Types.ObjectLockConfiguration{
		ObjectLockEnabled: s3Types.ObjectLockEnabledEnabled,
		Rule:              expandBucketLockConfigurationRule(d.Get("rule").([]interface{})),
	}

	input := &s3.PutObjectLockConfigurationInput{
		Bucket:                  aws.String(bucket),
		ObjectLockConfiguration: lockConfig,
	}

	_, err = conn.PutObjectLockConfiguration(ctx, input)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating Object bucket lock configuration (%s): %w", d.Id(), err))
	}

	return resourceObjectLockConfigurationRead(ctx, d, m)
}

func resourceObjectLockConfigurationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	conn, _, bucket, err := s3ClientWithRegionAndName(ctx, d, m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	input := &s3.PutObjectLockConfigurationInput{
		Bucket: aws.String(bucket),
		ObjectLockConfiguration: &s3Types.ObjectLockConfiguration{
			ObjectLockEnabled: s3Types.ObjectLockEnabledEnabled,
		},
	}

	_, err = conn.PutObjectLockConfiguration(ctx, input)

	if errors.As(err, new(*s3Types.NoSuchBucket)) {
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting Object bucket lock configuration (%s): %w", d.Id(), err))
	}

	return nil
}

func expandBucketLockConfigurationRule(l []interface{}) *s3Types.ObjectLockRule {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	return &s3Types.ObjectLockRule{
		DefaultRetention: expandBucketLockConfigurationRuleDefaultRetention(tfMap["default_retention"].([]interface{})),
	}
}

func expandBucketLockConfigurationRuleDefaultRetention(l []interface{}) *s3Types.DefaultRetention {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &s3Types.DefaultRetention{
		Mode: s3Types.ObjectLockRetentionMode(tfMap["mode"].(string)),
	}

	if v, ok := tfMap["days"].(int); ok && v > 0 {
		result.Days = aws.Int32(int32(v))
	}

	if v, ok := tfMap["years"].(int); ok && v > 0 {
		result.Years = aws.Int32(int32(v))
	}

	return result
}

func flattenBucketLockConfigurationRule(i *s3Types.ObjectLockRule) []interface{} {
	if i == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	m["default_retention"] = flattenBucketLockConfigurationRuleDefaultRetention(i.DefaultRetention)

	return []interface{}{m}
}

func flattenBucketLockConfigurationRuleDefaultRetention(i *s3Types.DefaultRetention) []interface{} {
	if i == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	m["mode"] = i.Mode

	if i.Days != nil {
		m["days"] = i.Days
	}

	if i.Years != nil {
		m["years"] = i.Years
	}

	return []interface{}{m}
}
