package scaleway

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
)

func ResourceScalewayObjectBucketPolicy() *schema.Resource {
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
				Type:             schema.TypeString,
				Required:         true,
				Description:      "The bucket's name or regional ID.",
				DiffSuppressFunc: diffSuppressFuncLocality,
			},
			"policy": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "The text of the policy.",
				DiffSuppressFunc: SuppressEquivalentPolicyDiffs,
			},
			"region":     regional.Schema(),
			"project_id": projectIDSchema(),
		},
	}
}

func resourceScalewayObjectBucketPolicyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	s3Client, region, err := s3ClientWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	regionalID := regional.ExpandID(d.Get("bucket"))
	bucket := regionalID.ID
	bucketRegion := regionalID.Region
	tflog.Debug(ctx, "bucket name: "+bucket)

	if bucketRegion != "" && bucketRegion != region {
		s3Client, err = s3ClientForceRegion(d, m, bucketRegion.String())
		if err != nil {
			return diag.FromErr(err)
		}
		region = bucketRegion
	}

	policy, err := structure.NormalizeJsonString(d.Get("policy").(string))
	if err != nil {
		return diag.FromErr(fmt.Errorf("policy (%s) is an invalid JSON: %w", policy, err))
	}

	tflog.Debug(ctx, fmt.Sprintf("[DEBUG] SCW bucket: %s, put policy: %s", bucket, policy))

	params := &s3.PutBucketPolicyInput{
		Bucket: scw.StringPtr(bucket),
		Policy: scw.StringPtr(policy),
	}

	err = retry.RetryContext(ctx, 1*time.Minute, func() *retry.RetryError {
		_, err := s3Client.PutBucketPolicyWithContext(ctx, params)
		if tfawserr.ErrCodeEquals(err, "MalformedPolicy") {
			return retry.RetryableError(err)
		}
		if err != nil {
			return retry.NonRetryableError(err)
		}
		return nil
	})
	if TimedOut(err) {
		_, err = s3Client.PutBucketPolicyWithContext(ctx, params)
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error putting SCW bucket policy: %s", err))
	}

	d.SetId(regional.NewIDString(region, bucket))

	return resourceScalewayObjectBucketPolicyRead(ctx, d, m)
}

//gocyclo:ignore
func resourceScalewayObjectBucketPolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	s3Client, region, _, err := s3ClientWithRegionAndName(d, m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	regionalID := regional.ExpandID(d.Id())
	bucket := regionalID.ID

	_ = d.Set("region", region)

	tflog.Debug(ctx, "[DEBUG] SCW bucket policy, read for bucket: "+d.Id())
	pol, err := s3Client.GetBucketPolicyWithContext(ctx, &s3.GetBucketPolicyInput{
		Bucket: aws.String(bucket),
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, ErrCodeNoSuchBucketPolicy, s3.ErrCodeNoSuchBucket) {
		tflog.Warn(ctx, fmt.Sprintf("[WARN] SCW Bucket Policy (%s) not found, removing from state", d.Id()))
		d.SetId("")
		return nil
	}

	v := ""
	if err == nil && pol.Policy != nil {
		v = aws.StringValue(pol.Policy)
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

	if err := d.Set("bucket", regionalID.String()); err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics
	acl, err := s3Client.GetBucketAclWithContext(ctx, &s3.GetBucketAclInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		if bucketFound, _ := addReadBucketErrorDiagnostic(&diags, err, "acl", ""); !bucketFound {
			return diags
		}
	} else if acl != nil && acl.Owner != nil {
		_ = d.Set("project_id", normalizeOwnerID(acl.Owner.ID))
	}

	return diags
}

func resourceScalewayObjectBucketPolicyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	s3Client, _, bucketName, err := s3ClientWithRegionAndName(d, m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	tflog.Debug(ctx, fmt.Sprintf("scw object bucket: %s, delete policy", bucketName))
	_, err = s3Client.DeleteBucketPolicy(&s3.DeleteBucketPolicyInput{
		Bucket: aws.String(bucketName),
	})

	if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting SCW Object policy: %s", err))
	}

	return nil
}
