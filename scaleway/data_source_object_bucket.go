package scaleway

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"log"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceScalewayObjectBucket() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayObjectBucket().Schema)

	// Set 'Optional' schema elements
	addOptionalFieldsToSchema(dsSchema, "name", "region")

	return &schema.Resource{
		ReadContext: dataSourceScalewayObjectStorageRead,
		Schema:      dsSchema,
	}
}

func dataSourceScalewayObjectStorageRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s3Client, region, err := s3ClientWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	bucket := d.Get("name").(string)

	input := &s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	}

	log.Printf("[DEBUG] Reading Object Storage bucket: %s", input)
	_, err = s3Client.HeadBucket(input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("failed getting Object Storage bucket (%s): %w", bucket, err))
	}

	bucketRegionalID := newRegionalIDString(region, bucket)
	d.SetId(bucketRegionalID)
	return resourceScalewayObjectBucketRead(ctx, d, meta)
}
