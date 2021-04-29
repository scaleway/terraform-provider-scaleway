package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceScalewayRDBDatabase() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayRdbDatabase().Schema)

	dsSchema["instance_id"].Computed = false
	dsSchema["instance_id"].Required = true
	addOptionalFieldsToSchema(dsSchema, "name")

	return &schema.Resource{
		ReadContext: dataSourceScalewayRDBDatabaseRead,
		Schema:      dsSchema,
	}
}

func dataSourceScalewayRDBDatabaseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	_, region, err := rdbAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	instanceID, _ := d.GetOk("instance_id")
	dbName, _ := d.GetOk("name")

	_, _, err = parseLocalizedID(instanceID.(string))
	regionalID := instanceID
	if err != nil {
		regionalID = datasourceNewRegionalizedID(instanceID, region)
	}

	d.SetId(fmt.Sprintf("%s/%s", regionalID, dbName.(string)))
	err = d.Set("instance_id", regionalID)
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceScalewayRdbDatabaseRead(ctx, d, meta)
}
