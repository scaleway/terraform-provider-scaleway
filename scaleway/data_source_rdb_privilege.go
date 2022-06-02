package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceScalewayRDBPrivilege() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayRdbPrivilege().Schema)

	fixDatasourceSchemaFlags(dsSchema, true, "instance_id", "user_name", "database_name")

	return &schema.Resource{
		ReadContext: dataSourceScalewayRDBPrivilegeRead,
		Schema:      dsSchema,
	}
}

func dataSourceScalewayRDBPrivilegeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	instanceID, _ := d.Get("instance_id").(string)

	d.SetId(instanceID)
	return resourceScalewayRdbPrivilegeRead(ctx, d, meta)
}
