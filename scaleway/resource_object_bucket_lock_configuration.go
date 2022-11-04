package scaleway

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	objectBucketLockConfigurationRetry = 5 * time.Second
)

func resourceObjectLockConfiguration() *schema.Resource {
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
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 63),
				Description:  "The bucket name.",
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
										ConflictsWith: []string{"rule.default_retention.years"},
									},
									"years": {
										Type:          schema.TypeInt,
										Optional:      true,
										Description:   "The number of years that you want to specify for the default retention period.",
										ConflictsWith: []string{"rule.default_retention.days"},
									},
								},
							},
						},
					},
				},
				Description: "Specifies the Object Lock rule for the specified object.",
			},
		},
	}
}

func resourceObjectLockConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn, region, err := s3ClientWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	bucket := expandID(d.Get("bucket").(string))

	input := &s3.PutObjectLockConfigurationInput{
		Bucket: aws.String(bucket),
		ObjectLockConfiguration: &s3.ObjectLockConfiguration{
			ObjectLockEnabled: aws.String("Enabled"),
			Rule:              expandBucketLockConfigurationRule(d.Get("rule").([]interface{})),
		},
	}

	_, err = RetryWhenAWSErrCodeEqualsContext(ctx, objectBucketLockConfigurationRetry, func() (interface{}, error) {
		return conn.PutObjectLockConfigurationWithContext(ctx, input)
	}, s3.ErrCodeNoSuchBucket)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating object bucket (%s) lock configuration: %w", bucket, err))
	}

	d.SetId(newRegionalIDString(region, bucket))

	return resourceObjectLockConfigurationRead(ctx, d, meta)
}

func resourceObjectLockConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn, _, bucket, err := s3ClientWithRegionAndName(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	input := &s3.GetObjectLockConfigurationInput{
		Bucket: aws.String(bucket),
	}

	output, err := conn.GetObjectLockConfigurationWithContext(ctx, input)
	if !d.IsNewResource() && ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
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

	_ = d.Set("bucket", bucket)
	_ = d.Set("rule", flattenBucketLockConfigurationRule(output.ObjectLockConfiguration.Rule))

	return nil
}

func resourceObjectLockConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn, _, bucket, err := s3ClientWithRegionAndName(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	lockConfig := &s3.ObjectLockConfiguration{
		ObjectLockEnabled: aws.String(s3.ObjectLockEnabledEnabled),
		Rule:              expandBucketLockConfigurationRule(d.Get("rule").([]interface{})),
	}

	input := &s3.PutObjectLockConfigurationInput{
		Bucket:                  aws.String(bucket),
		ObjectLockConfiguration: lockConfig,
	}

	_, err = conn.PutObjectLockConfigurationWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating Object bucket lock configuration (%s): %w", d.Id(), err))
	}

	return resourceObjectLockConfigurationRead(ctx, d, meta)
}

func resourceObjectLockConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn, _, bucket, err := s3ClientWithRegionAndName(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	input := &s3.PutObjectLockConfigurationInput{
		Bucket: aws.String(bucket),
		ObjectLockConfiguration: &s3.ObjectLockConfiguration{
			ObjectLockEnabled: aws.String(s3.ObjectLockEnabledEnabled),
		},
	}

	_, err = conn.PutObjectLockConfigurationWithContext(ctx, input)

	if ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting Object bucket lock configuration (%s): %w", d.Id(), err))
	}

	return nil
}

func expandBucketLockConfigurationRule(l []interface{}) *s3.ObjectLockRule {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	return &s3.ObjectLockRule{
		DefaultRetention: expandBucketLockConfigurationRuleDefaultRetention(tfMap["default_retention"].([]interface{})),
	}
}

func expandBucketLockConfigurationRuleDefaultRetention(l []interface{}) *s3.DefaultRetention {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &s3.DefaultRetention{
		Mode: aws.String(tfMap["mode"].(string)),
	}

	if v, ok := tfMap["days"].(int); ok && v > 0 {
		result.Days = aws.Int64(int64(v))
	}

	if v, ok := tfMap["years"].(int); ok && v > 0 {
		result.Years = aws.Int64(int64(v))
	}

	return result
}

func flattenBucketLockConfigurationRule(i *s3.ObjectLockRule) []interface{} {
	if i == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	m["default_retention"] = flattenBucketLockConfigurationRuleDefaultRetention(i.DefaultRetention)

	return []interface{}{m}
}

func flattenBucketLockConfigurationRuleDefaultRetention(i *s3.DefaultRetention) []interface{} {
	if i == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	m["mode"] = aws.StringValue(i.Mode)

	if i.Days != nil {
		m["days"] = aws.Int64Value(i.Days)
	}

	if i.Years != nil {
		m["years"] = aws.Int64Value(i.Years)
	}

	return []interface{}{m}
}
