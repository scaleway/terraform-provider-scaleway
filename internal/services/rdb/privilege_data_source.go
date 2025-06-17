package rdb

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
)

func DataSourcePrivilege() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasource.SchemaFromResourceSchema(ResourcePrivilege().Schema)

	datasource.FixDatasourceSchemaFlags(dsSchema, true, "instance_id", "user_name", "database_name")

	// Set 'Optional' schema elements
	datasource.AddOptionalFieldsToSchema(dsSchema, "region")

	return &schema.Resource{
		ReadContext: DataSourceRDBPrivilegeRead,
		Schema:      dsSchema,
	}
}

// DataSourceRDBPrivilegeRead
func DataSourceRDBPrivilegeRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	_, region, err := newAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	instanceID := locality.ExpandID(d.Get("instance_id").(string))
	userName, _ := d.Get("user_name").(string)
	databaseName, _ := d.Get("database_name").(string)

	d.SetId(ResourceRdbUserPrivilegeID(region, instanceID, databaseName, userName))

	return ResourceRdbPrivilegeRead(ctx, d, m)
}
