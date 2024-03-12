package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

func dataSourceScalewayRDBPrivilege() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayRdbPrivilege().Schema)

	fixDatasourceSchemaFlags(dsSchema, true, "instance_id", "user_name", "database_name")

	// Set 'Optional' schema elements
	addOptionalFieldsToSchema(dsSchema, "region")
	return &schema.Resource{
		ReadContext: dataSourceScalewayRDBPrivilegeRead,
		Schema:      dsSchema,
	}
}

// dataSourceScalewayRDBPrivilegeRead
func dataSourceScalewayRDBPrivilegeRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	_, region, err := rdbAPIWithRegion(d, m.(*meta.Meta))
	if err != nil {
		return diag.FromErr(err)
	}

	instanceID := locality.ExpandID(d.Get("instance_id").(string))
	userName, _ := d.Get("user_name").(string)
	databaseName, _ := d.Get("database_name").(string)

	d.SetId(resourceScalewayRdbUserPrivilegeID(region, instanceID, databaseName, userName))
	return resourceScalewayRdbPrivilegeRead(ctx, d, m)
}
