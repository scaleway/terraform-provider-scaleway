package object

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
)

//go:embed descriptions/bucket_server_side_encryption_configuration_data_source.md
var bucketServerSideEncryptionConfigurationDataSourceDescription string

func DataSourceBucketServerSideEncryptionConfiguration() *schema.Resource {
	dsSchema := datasource.SchemaFromResourceSchema(ResourceBucketServerSideEncryptionConfiguration().SchemaFunc())

	filterFields := []string{"bucket"}

	dsSchema["bucket_server_side_encryption_configuration_id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The ID of the bucket server side encryption configuration",
		ConflictsWith: filterFields,
	}

	datasource.AddOptionalFieldsToSchema(dsSchema, "bucket")

	dsSchema["bucket"].ConflictsWith = []string{"bucket_server_side_encryption_configuration_id"}

	return &schema.Resource{
		ReadContext: DataSourceBucketServerSideEncryptionConfigurationRead,
		Description: bucketServerSideEncryptionConfigurationDataSourceDescription,
		Schema:      dsSchema,
	}
}

func DataSourceBucketServerSideEncryptionConfigurationRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	configID, idExists := d.GetOk("bucket_server_side_encryption_configuration_id")
	if idExists {
		return dataSourceBucketServerSideEncryptionConfigurationReadByID(ctx, d, m, configID.(string))
	}

	return dataSourceBucketServerSideEncryptionConfigurationReadByFilters(ctx, d, m)
}

func dataSourceBucketServerSideEncryptionConfigurationReadByID(ctx context.Context, d *schema.ResourceData, m any, configID string) diag.Diagnostics {
	s3Client, region, bucketName, err := s3ClientWithRegionAndName(ctx, d, m, configID)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(configID)

	sse, err := findServerSideEncryptionConfiguration(ctx, s3Client, bucketName)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(regional.NewIDString(region, bucketName))

	_ = d.Set("bucket", bucketName)
	if err := d.Set("rule", flattenServerSideEncryptionRules(sse.Rules)); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func dataSourceBucketServerSideEncryptionConfigurationReadByFilters(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	bucketName, ok := d.GetOk("bucket")
	if !ok {
		return diag.FromErr(errors.New("bucket is required when bucket_server_side_encryption_configuration_id is not specified"))
	}

	s3Client, region, err := s3ClientWithRegion(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	sse, err := findServerSideEncryptionConfiguration(ctx, s3Client, bucketName.(string))
	if err != nil {
		return diag.FromErr(fmt.Errorf("no server side encryption configuration found for bucket %s: %w", bucketName, err))
	}

	d.SetId(regional.NewIDString(region, bucketName.(string)))

	_ = d.Set("bucket", bucketName)
	if err := d.Set("rule", flattenServerSideEncryptionRules(sse.Rules)); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
