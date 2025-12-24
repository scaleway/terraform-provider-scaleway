package object

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
)

func ResourceBucketWebsiteConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBucketWebsiteConfigurationCreate,
		ReadContext:   resourceBucketWebsiteConfigurationRead,
		UpdateContext: resourceBucketWebsiteConfigurationUpdate,
		DeleteContext: resourceBucketWebsiteConfigurationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaFunc: bucketWebsiteConfigurationSchema,
		Identity:   identity.DefaultRegional(),
	}
}

func bucketWebsiteConfigurationSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"bucket": {
			Type:             schema.TypeString,
			Required:         true,
			ForceNew:         true,
			ValidateFunc:     validation.StringLenBetween(1, 63),
			Description:      "The bucket's name or regional ID.",
			DiffSuppressFunc: dsf.Locality,
		},
		"index_document": {
			Type:     schema.TypeList,
			Required: true,
			MinItems: 1,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"suffix": {
						Type:        schema.TypeString,
						Description: "Suffix that will be added to the index.",
						Required:    true,
					},
				},
			},
			Description: "The name of the index document for the website.",
		},
		"error_document": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"key": {
						Type:        schema.TypeString,
						Description: "Key for the object to use as an error document.",
						Required:    true,
					},
				},
			},
			Description: "The name of the error document for the website.",
		},
		"website_endpoint": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The domain of the website endpoint.",
		},
		"website_domain": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The website endpoint.",
		},
		"region":     regional.Schema(),
		"project_id": account.ProjectIDSchema(),
	}
}

func resourceBucketWebsiteConfigurationCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
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

	websiteConfig := &s3Types.WebsiteConfiguration{
		IndexDocument: expandBucketWebsiteConfigurationIndexDocument(d.Get("index_document").([]any)),
	}

	if v, ok := d.GetOk("error_document"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		websiteConfig.ErrorDocument = expandBucketWebsiteConfigurationErrorDocument(v.([]any))
	}

	_, err = conn.ListObjects(ctx, &s3.ListObjectsInput{
		Bucket: scw.StringPtr(bucket),
	})
	if err != nil {
		return diag.FromErr(fmt.Errorf("couldn't read bucket: %w", err))
	}

	input := &s3.PutBucketWebsiteInput{
		Bucket:               aws.String(bucket),
		WebsiteConfiguration: websiteConfig,
	}

	_, err = conn.PutBucketWebsite(ctx, input)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating object bucket (%s) website configuration: %w", bucket, err))
	}

	d.SetId(regional.NewIDString(region, bucket))

	return resourceBucketWebsiteConfigurationRead(ctx, d, m)
}

func resourceBucketWebsiteConfigurationRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	conn, region, bucket, err := s3ClientWithRegionAndName(ctx, d, m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	input := &s3.GetBucketWebsiteInput{
		Bucket: aws.String(bucket),
	}

	// expectedBucketOwner and routing not supported

	_, err = conn.ListObjects(ctx, &s3.ListObjectsInput{
		Bucket: scw.StringPtr(bucket),
	})
	if err != nil {
		if IsS3Err(err, ErrCodeNoSuchBucket, "") {
			tflog.Error(ctx, fmt.Sprintf("Bucket %q was not found - removing from state!", bucket))
			d.SetId("")

			return nil
		}

		return diag.FromErr(fmt.Errorf("couldn't read bucket: %w", err))
	}

	output, err := conn.GetBucketWebsite(ctx, input)
	if !d.IsNewResource() && ErrCodeEquals(err, ErrCodeNoSuchBucket, ErrCodeNoSuchWebsiteConfiguration) {
		tflog.Debug(ctx, fmt.Sprintf("[WARN] Object Bucket Website Configuration (%s) not found, removing from state", d.Id()))
		d.SetId("")

		return nil
	}

	if output == nil {
		if d.IsNewResource() {
			return diag.FromErr(fmt.Errorf("error reading object bucket website configuration (%s): empty output", d.Id()))
		}

		tflog.Info(ctx, fmt.Sprintf("[WARN] object Bucket Website Configuration (%s) not found, removing from state", d.Id()))
		d.SetId("")

		return nil
	}

	_ = d.Set("bucket", bucket)
	_ = d.Set("index_document", flattenBucketWebsiteConfigurationIndexDocument(output.IndexDocument))

	if err := d.Set("error_document", flattenBucketWebsiteConfigurationErrorDocument(output.ErrorDocument)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting error_document: %w", err))
	}

	websiteEndpoint := WebsiteEndpoint(bucket, region)

	if websiteEndpoint != nil {
		_ = d.Set("website_endpoint", websiteEndpoint.Endpoint)
		_ = d.Set("website_domain", websiteEndpoint.Domain)
	}

	acl, err := conn.GetBucketAcl(ctx, &s3.GetBucketAclInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		return diag.FromErr(fmt.Errorf("couldn't read bucket acl: %w", err))
	}

	_ = d.Set("project_id", NormalizeOwnerID(acl.Owner.ID))

	return nil
}

func resourceBucketWebsiteConfigurationUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	conn, _, bucket, err := s3ClientWithRegionAndName(ctx, d, m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	websiteConfig := &s3Types.WebsiteConfiguration{
		IndexDocument: expandBucketWebsiteConfigurationIndexDocument(d.Get("index_document").([]any)),
	}

	if v, ok := d.GetOk("error_document"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		websiteConfig.ErrorDocument = expandBucketWebsiteConfigurationErrorDocument(v.([]any))
	}

	input := &s3.PutBucketWebsiteInput{
		Bucket:               aws.String(bucket),
		WebsiteConfiguration: websiteConfig,
	}

	_, err = conn.PutBucketWebsite(ctx, input)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating Object bucket website configuration (%s): %w", d.Id(), err))
	}

	return resourceBucketWebsiteConfigurationRead(ctx, d, m)
}

func resourceBucketWebsiteConfigurationDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	conn, _, bucket, err := s3ClientWithRegionAndName(ctx, d, m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	input := &s3.DeleteBucketWebsiteInput{
		Bucket: aws.String(bucket),
	}

	_, err = conn.DeleteBucketWebsite(ctx, input)

	if ErrCodeEquals(err, ErrCodeNoSuchBucket, ErrCodeNoSuchWebsiteConfiguration) {
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting Object bucket website configuration (%s): %w", d.Id(), err))
	}

	return nil
}

func expandBucketWebsiteConfigurationErrorDocument(l []any) *s3Types.ErrorDocument {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]any)
	if !ok {
		return nil
	}

	result := &s3Types.ErrorDocument{}

	if v, ok := tfMap["key"].(string); ok && v != "" {
		result.Key = aws.String(v)
	}

	return result
}

func expandBucketWebsiteConfigurationIndexDocument(l []any) *s3Types.IndexDocument {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]any)
	if !ok {
		return nil
	}

	result := &s3Types.IndexDocument{}

	if v, ok := tfMap["suffix"].(string); ok && v != "" {
		result.Suffix = aws.String(v)
	}

	return result
}

func flattenBucketWebsiteConfigurationIndexDocument(i *s3Types.IndexDocument) []any {
	if i == nil {
		return []any{}
	}

	m := make(map[string]any)

	if i.Suffix != nil {
		m["suffix"] = aws.ToString(i.Suffix)
	}

	return []any{m}
}

func flattenBucketWebsiteConfigurationErrorDocument(e *s3Types.ErrorDocument) []any {
	if e == nil {
		return []any{}
	}

	m := make(map[string]any)

	if e.Key != nil {
		m["key"] = aws.ToString(e.Key)
	}

	return []any{m}
}
