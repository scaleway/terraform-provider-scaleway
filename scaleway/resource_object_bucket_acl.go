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
			"acl": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "private",
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
	acl := d.Get("acl").(string)

	input := &s3.PutBucketAclInput{
		Bucket: aws.String(bucket),
	}

	if acl != "" {
		input.ACL = aws.String(acl)
	}

	_, err = retryOnAWSCode(ctx, s3.ErrCodeNoSuchBucket, func() (interface{}, error) {
		return conn.PutBucketAclWithContext(ctx, input)
	})
	if err != nil {
		return diag.FromErr(fmt.Errorf("error putting Object Storage ACL: %s", err))
	}

	d.SetId(BucketACLCreateResourceID(region, bucket, acl))

	return resourceBucketACLRead(ctx, d, meta)
}

func resourceBucketACLRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	_, region, bucket, acl, err := s3ClientWithRegionWithNameACL(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		tflog.Info(ctx, fmt.Sprintf("[WARN] Object Bucket ACL (%s) not found, removing from state", d.Id()))
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting object storage bucket ACL (%s): %w", d.Id(), err))
	}

	_ = d.Set("acl", acl)
	_ = d.Set("region", region)
	_ = d.Set("bucket", expandID(bucket))

	return nil
}

// BucketACLCreateResourceID is a method for creating an ID string
// with the bucket name and/or ACL.
func BucketACLCreateResourceID(region scw.Region, bucket, acl string) string {
	if acl == "" {
		return bucket
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
