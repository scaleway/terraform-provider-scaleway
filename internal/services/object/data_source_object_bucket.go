package object

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
)

func DataSourceBucket() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasource.SchemaFromResourceSchema(ResourceBucket().Schema)

	// Set 'Optional' schema elements
	datasource.AddOptionalFieldsToSchema(dsSchema, "name", "region", "project_id")

	return &schema.Resource{
		ReadContext: DataSourceObjectStorageRead,
		Schema:      dsSchema,
	}
}

func DataSourceObjectStorageRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	s3Client, region, err := s3ClientWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	regionalID := regional.ExpandID(d.Get("name"))
	bucket := regionalID.ID
	bucketRegion := regionalID.Region

	if bucketRegion != "" && bucketRegion != region {
		s3Client, err = s3ClientForceRegion(d, m, bucketRegion.String())
		if err != nil {
			return diag.FromErr(err)
		}
		region = bucketRegion
	}

	input := &s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	}

	log.Printf("[DEBUG] Reading Object Storage bucket: %s", input)
	_, err = s3Client.HeadBucketWithContext(ctx, input)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed getting Object Storage bucket (%s): %w", bucket, err))
	}

	acl, err := s3Client.GetBucketAclWithContext(ctx, &s3.GetBucketAclInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		return diag.FromErr(fmt.Errorf("couldn't read bucket acl: %s", err))
	}
	_ = d.Set("project_id", normalizeOwnerID(acl.Owner.ID))

	bucketRegionalID := regional.NewIDString(region, bucket)
	d.SetId(bucketRegionalID)
	return resourceObjectBucketRead(ctx, d, m)
}
