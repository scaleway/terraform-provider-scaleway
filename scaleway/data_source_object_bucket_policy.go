package scaleway

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
)

func dataSourceScalewayObjectBucketPolicy() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayObjectBucketPolicy().Schema)

	fixDatasourceSchemaFlags(dsSchema, true, "bucket")
	addOptionalFieldsToSchema(dsSchema, "region", "project_id")

	return &schema.Resource{
		ReadContext: dataSourceScalewayObjectBucketPolicyRead,
		Schema:      dsSchema,
	}
}

func dataSourceScalewayObjectBucketPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s3Client, region, err := s3ClientWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	regionalID := expandRegionalID(d.Get("bucket"))
	bucket := regionalID.ID
	bucketRegion := regionalID.Region
	tflog.Debug(ctx, fmt.Sprintf("bucket name: %s", bucket))

	if bucketRegion != "" && bucketRegion != region {
		s3Client, err = s3ClientForceRegion(d, meta, bucketRegion.String())
		if err != nil {
			return diag.FromErr(err)
		}
		region = bucketRegion
	}
	_ = d.Set("region", region)

	tflog.Debug(ctx, fmt.Sprintf("[DEBUG] SCW bucket policy, read for bucket: %s", d.Id()))
	policy, err := s3Client.GetBucketPolicyWithContext(ctx, &s3.GetBucketPolicyInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		if tfawserr.ErrCodeEquals(err, ErrCodeNoSuchBucketPolicy, s3.ErrCodeNoSuchBucket) {
			return diag.FromErr(fmt.Errorf("bucket %s doesn't exist or has no policy", bucket))
		}

		return diag.FromErr(fmt.Errorf("couldn't read bucket %s policy: %s", bucket, err))
	}

	policyString := "{}"
	if err == nil && policy.Policy != nil {
		policyString = aws.StringValue(policy.Policy)
	}

	policyJSON, err := structure.NormalizeJsonString(policyString)
	if err != nil {
		return diag.FromErr(fmt.Errorf("policy (%s) is an invalid JSON: %w", policyString, err))
	}

	_ = d.Set("policy", policyJSON)

	acl, err := s3Client.GetBucketAclWithContext(ctx, &s3.GetBucketAclInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		return diag.FromErr(fmt.Errorf("couldn't read bucket acl: %s", err))
	}
	_ = d.Set("project_id", normalizeOwnerID(acl.Owner.ID))

	d.SetId(newRegionalIDString(region, bucket))
	return nil
}
