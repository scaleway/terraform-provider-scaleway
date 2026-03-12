package object

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
)

func DataSourceObject() *schema.Resource {
	dsSchema := datasource.SchemaFromResourceSchema(ResourceObject().SchemaFunc())

	datasource.FixDatasourceSchemaFlags(dsSchema, true, "bucket", "key")

	datasource.AddOptionalFieldsToSchema(dsSchema, "region", "project_id")

	return &schema.Resource{
		ReadContext: DataSourceObjectRead,
		Schema:      dsSchema,
	}
}

func DataSourceObjectRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	s3Client, region, err := s3ClientWithRegion(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	regionalID := regional.ExpandID(d.Get("bucket"))
	bucket := regionalID.ID
	bucketRegion := regionalID.Region

	if bucketRegion != "" && bucketRegion != region {
		s3Client, err = s3ClientForceRegion(ctx, d, m, bucketRegion.String())
		if err != nil {
			return diag.FromErr(err)
		}

		region = bucketRegion
	}

	_ = d.Set("region", region)

	key := d.Get("key").(string)

	tflog.Debug(ctx, fmt.Sprintf("SCW object read for bucket=%s key=%s", bucket, key))

	_, err = s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return diag.FromErr(fmt.Errorf("couldn't read object %s/%s: %w", bucket, key, err))
	}

	d.SetId(regional.NewIDString(region, objectID(bucket, key)))

	return resourceObjectRead(ctx, d, m)
}
