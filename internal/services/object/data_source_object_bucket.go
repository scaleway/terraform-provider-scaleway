package object

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
)

func DataSourceBucket() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasource.SchemaFromResourceSchema(ResourceBucket().SchemaFunc())

	// Set 'Optional' schema elements
	datasource.AddOptionalFieldsToSchema(dsSchema, "name", "region", "project_id")

	return &schema.Resource{
		ReadContext: DataSourceObjectStorageRead,
		Schema:      dsSchema,
		Identity:    identity.DefaultRegional(),
	}
}

func DataSourceObjectStorageRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	s3Client, region, err := s3ClientWithRegion(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	regionalID := regional.ExpandID(d.Get("name"))
	bucket := regionalID.ID
	bucketRegion := regionalID.Region

	if bucketRegion != "" && bucketRegion != region {
		s3Client, err = s3ClientForceRegion(ctx, d, m, bucketRegion.String())
		if err != nil {
			return diag.FromErr(err)
		}

		region = bucketRegion
	}

	input := &s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	}

	log.Printf("[DEBUG] Reading Object Storage bucket: %s", bucket)

	_, err = s3Client.HeadBucket(ctx, input)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed getting Object Storage bucket (%s): %w", bucket, err))
	}

	var diags diag.Diagnostics

	projectId, diags, ok := setProjectId(ctx, d, bucket, s3Client, &diags)
	if !ok {
		return diags
	}

	bucketRegionalID := regional.NewIDString(region, bucket)
	d.SetId(bucketRegionalID)

	return dataSourceObjectBucketRead(ctx, d, m, s3Client, bucket, region, projectId)
}

// dataSourceObjectBucketRead reads bucket attributes for the datasource without manipulating Identity
func dataSourceObjectBucketRead(ctx context.Context, d *schema.ResourceData, m any, s3Client *s3.Client, bucketName string, region scw.Region, projectId string) diag.Diagnostics {
	var diags diag.Diagnostics

	_ = d.Set("name", bucketName)
	_ = d.Set("region", region.String())
	_ = d.Set("project_id", projectId)

	// Read tags
	tagsResponse, err := s3Client.GetBucketTagging(ctx, &s3.GetBucketTaggingInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil && !IsS3Err(err, ErrCodeNoSuchTagSet, "") {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  fmt.Sprintf("Couldn't read bucket tags: %s", err),
		})
	} else if err == nil {
		_ = d.Set("tags", flattenObjectBucketTags(tagsResponse.TagSet))
	}

	// Read versioning
	versioningResponse, err := s3Client.GetBucketVersioning(ctx, &s3.GetBucketVersioningInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  fmt.Sprintf("Couldn't read bucket versioning: %s", err),
		})
	} else {
		_ = d.Set("versioning", FlattenObjectBucketVersioning(versioningResponse))
	}

	// Read CORS
	corsResponse, err := s3Client.GetBucketCors(ctx, &s3.GetBucketCorsInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil && !IsS3Err(err, ErrCodeNoSuchCORSConfiguration, "") {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  fmt.Sprintf("Couldn't read bucket CORS: %s", err),
		})
	} else {
		_ = d.Set("cors_rule", flattenBucketCORS(corsResponse))
	}

	// Read lifecycle configuration
	lifecycle, err := s3Client.GetBucketLifecycleConfiguration(ctx, &s3.GetBucketLifecycleConfigurationInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil && !IsS3Err(err, ErrCodeNoSuchLifecycleConfiguration, "") {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  fmt.Sprintf("Couldn't read bucket lifecycle: %s", err),
		})
	} else {
		lifecycleRules := resourceBucketLifecycleRulesRead(lifecycle, d)
		if err := d.Set("lifecycle_rule", lifecycleRules); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("error setting lifecycle_rule: %s", err),
			})
		}
	}

	// Read endpoint URLs
	_ = d.Set("endpoint", objectBucketEndpointURL(bucketName, region))
	_ = d.Set("api_endpoint", objectBucketAPIEndpointURL(region))

	return diags
}
