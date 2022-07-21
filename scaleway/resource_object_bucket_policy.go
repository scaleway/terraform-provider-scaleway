package scaleway

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceScalewayObjectBucketPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBucketPolicyPut,
		ReadContext:   resourceBucketPolicyRead,
		UpdateContext: resourceBucketPolicyPut,
		DeleteContext: resourceBucketPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"policy": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: SuppressEquivalentPolicyDiffs,
			},
		},
	}
}

func resourceBucketPolicyPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s3Client, _, err := s3ClientWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	bucket := d.Get("bucket").(string)

	policy, err := structure.NormalizeJsonString(d.Get("policy").(string))
	if err != nil {
<<<<<<< HEAD
		return diag.Errorf("policy (%s) is an invalid JSON: %s", policy, err)
=======
		return diag.Errorf("policy (%s) is an invalid JSON: %w", policy, err)
>>>>>>> e0eb24f7 (feat(object): add support for bucket policy)
	}

	tflog.Debug(ctx, fmt.Sprintf("S3 bucket: %s, put policy: %s", bucket, policy))

	params := &s3.PutBucketPolicyInput{
		Bucket: aws.String(bucket),
		Policy: aws.String(policy),
	}

	err = resource.RetryContext(ctx, 1*time.Minute, func() *resource.RetryError {
		_, err := s3Client.PutBucketPolicy(params)
		if tfawserr.ErrCodeEquals(err, "MalformedPolicy") {
			return resource.RetryableError(err)
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	})
<<<<<<< HEAD
	if err != nil {
		return diag.FromErr(err)
	}
=======
>>>>>>> e0eb24f7 (feat(object): add support for bucket policy)
	_, err = s3Client.PutBucketPolicy(params)
	if err != nil {
		return diag.Errorf("Error putting S3 policy: %s", err)
	}

	d.SetId(bucket)

	return nil
}

func resourceBucketPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s3Client, _, err := s3ClientWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	tflog.Debug(ctx, fmt.Sprintf("S3 bucket policy, read for bucket: %s", d.Id()))
	pol, err := s3Client.GetBucketPolicy(&s3.GetBucketPolicyInput{
		Bucket: aws.String(d.Id()),
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, ErrCodeNoSuchBucketPolicy, s3.ErrCodeNoSuchBucket) {
		tflog.Warn(ctx, fmt.Sprintf("S3 Bucket Policy (%s) not found, removing from state", d.Id()))
		d.SetId("")
		return nil
	}

	v := ""
	if err == nil && pol.Policy != nil {
		v = aws.StringValue(pol.Policy)
	}

	policyToSet, err := SecondJSONUnlessEquivalent(d.Get("policy").(string), v)
	if err != nil {
<<<<<<< HEAD
		return diag.Errorf("while setting policy (%s), encountered: %s", policyToSet, err)
=======
		return diag.Errorf("while setting policy (%s), encountered: %w", policyToSet, err)
>>>>>>> e0eb24f7 (feat(object): add support for bucket policy)
	}

	policyToSet, err = structure.NormalizeJsonString(policyToSet)

	if err != nil {
<<<<<<< HEAD
		return diag.Errorf("policy (%s) is an invalid JSON: %s", policyToSet, err)
=======
		return diag.Errorf("policy (%s) is an invalid JSON: %w", policyToSet, err)
>>>>>>> e0eb24f7 (feat(object): add support for bucket policy)
	}

	if err := d.Set("policy", policyToSet); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("bucket", d.Id()); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceBucketPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s3Client, _, err := s3ClientWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	bucket := d.Get("bucket").(string)

	tflog.Debug(ctx, fmt.Sprintf("S3 bucket: %s, delete policy", bucket))
	_, err = s3Client.DeleteBucketPolicy(&s3.DeleteBucketPolicyInput{
		Bucket: aws.String(bucket),
	})

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "NoSuchBucket" {
			return nil
		}
		return diag.Errorf("Error deleting S3 policy: %s", err)
	}

	return nil
}
