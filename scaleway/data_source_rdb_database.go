package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
)

func DataSourceScalewayRDBDatabase() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasource.SchemaFromResourceSchema(ResourceScalewayRdbDatabase().Schema)

	datasource.FixDatasourceSchemaFlags(dsSchema, true, "instance_id", "name")

	// Set 'Optional' schema elements
	datasource.AddOptionalFieldsToSchema(dsSchema, "region")
	return &schema.Resource{
		ReadContext: dataSourceScalewayRDBDatabaseRead,
		Schema:      dsSchema,
	}
}

func dataSourceScalewayRDBDatabaseRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	_, region, err := rdbAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	instanceID, _ := d.GetOk("instance_id")
	dbName, _ := d.GetOk("name")

	_, _, err = locality.ParseLocalizedID(instanceID.(string))
	if err != nil {
		instanceID = datasource.NewRegionalID(instanceID, region)
	}

	d.SetId(fmt.Sprintf("%s/%s", instanceID, dbName.(string)))
	err = d.Set("instance_id", instanceID)
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceScalewayRdbDatabaseRead(ctx, d, m)
}
