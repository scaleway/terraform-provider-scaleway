package scaleway

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayObjectBucketPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayObjectBucketPolicyCreate,
		ReadContext:   resourceScalewayObjectBucketPolicyRead,
		UpdateContext: resourceScalewayObjectBucketPolicyCreate,
		DeleteContext: resourceScalewayObjectBucketPolicyDelete,
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(defaultObjectBucketTimeout),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The bucket name.",
			},
			"policy": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "The text of the policy.",
				DiffSuppressFunc: SuppressEquivalentPolicyDiffs,
			},
			"region": regionSchema(),
		},
	}
}

func resourceScalewayObjectBucketPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s3Client, region, err := s3ClientWithRegion(ctx, d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	bucket := expandID(d.Get("bucket"))
	tflog.Debug(ctx, fmt.Sprintf("bucket name: %s", bucket))

	policy, err := structure.NormalizeJsonString(d.Get("policy").(string))
	if err != nil {
		return diag.FromErr(fmt.Errorf("policy (%s) is an invalid JSON: %w", policy, err))
	}

	tflog.Debug(ctx, fmt.Sprintf("[DEBUG] SCW bucket: %s, put policy: %s", bucket, policy))

	params := &s3.PutBucketPolicyInput{
		Bucket: scw.StringPtr(bucket),
		Policy: scw.StringPtr(policy),
	}

	err = resource.RetryContext(ctx, 1*time.Minute, func() *resource.RetryError {
		_, err := s3Client.PutBucketPolicy(ctx, params)
		if isS3ErrCode(err, "MalformedPolicy", "") {
			return resource.RetryableError(err)
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if TimedOut(err) {
		_, err = s3Client.PutBucketPolicy(ctx, params)
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error putting SCW bucket policy: %s", err))
	}

	d.SetId(newRegionalIDString(region, bucket))

	return resourceScalewayObjectBucketPolicyRead(ctx, d, meta)
}

//gocyclo:ignore
func resourceScalewayObjectBucketPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s3Client, region, _, err := s3ClientWithRegionAndName(ctx, meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("region", region)

	tflog.Debug(ctx, fmt.Sprintf("[DEBUG] SCW bucket policy, read for bucket: %s", d.Id()))
	pol, err := s3Client.GetBucketPolicy(ctx, &s3.GetBucketPolicyInput{
		Bucket: aws.String(expandID(d.Id())),
	})

	if !d.IsNewResource() && isS3Err(err, &s3types.NoSuchBucket{}) || isS3ErrCode(err, ErrCodeNoSuchBucketPolicy, "") {
		tflog.Warn(ctx, fmt.Sprintf("[WARN] SCW Bucket Policy (%s) not found, removing from state", d.Id()))
		d.SetId("")
		return nil
	}

	v := ""
	if err == nil && pol.Policy != nil {
		v = *pol.Policy
	}

	policyToSet, err := SecondJSONUnlessEquivalent(d.Get("policy").(string), v)
	if err != nil {
		return diag.FromErr(fmt.Errorf("while setting policy (%s), encountered: %w", policyToSet, err))
	}

	policyToSet, err = structure.NormalizeJsonString(policyToSet)

	if err != nil {
		return diag.FromErr(fmt.Errorf("policy (%s) is an invalid JSON: %w", policyToSet, err))
	}

	if err := d.Set("policy", policyToSet); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("bucket", expandID(d.Id())); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceScalewayObjectBucketPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s3Client, _, bucketName, err := s3ClientWithRegionAndName(ctx, meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	tflog.Debug(ctx, fmt.Sprintf("scw object bucket: %s, delete policy", bucketName))
	_, err = s3Client.DeleteBucketPolicy(ctx, &s3.DeleteBucketPolicyInput{
		Bucket: aws.String(bucketName),
	})

	if isS3Err(err, &s3types.NoSuchBucket{}) {
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting SCW Object policy: %s", err))
	}

	return nil
}
