package scaleway

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
)

func dataSourceScalewayObjectBucketPolicy() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceBucketPolicyRead,

		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:     schema.TypeString,
				Required: true,
			},
			"policy": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceBucketPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s3Client, _, err := s3ClientWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	name := d.Get("bucket").(string)

	out, err := FindBucketPolicy(ctx, s3Client, name)
	if err != nil {
		return diag.Errorf("failed getting S3 bucket policy (%s): %s", name, err)
	}

	policy, err := structure.NormalizeJsonString(aws.StringValue(out.Policy))
	if err != nil {
		return diag.Errorf("policy (%s) is an invalid JSON: %s", policy, err)
	}

	d.SetId(name)
	_ = d.Set("policy", policy)

	return nil
}

func FindBucketPolicy(ctx context.Context, conn *s3.S3, name string) (*s3.GetBucketPolicyOutput, error) {
	in := &s3.GetBucketPolicyInput{
		Bucket: aws.String(name),
	}
	tflog.Debug(ctx, fmt.Sprintf("Reading S3 bucket policy: %s", in))

	out, err := conn.GetBucketPolicyWithContext(ctx, in)

	if ErrCodeEquals(err, ErrCodeNoSuchBucketPolicy) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	return out, nil
}
