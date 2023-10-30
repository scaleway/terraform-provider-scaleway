package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceScalewayDocumentDBDatabase() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayDocumentDBDatabase().Schema)

	addOptionalFieldsToSchema(dsSchema, "name", "region")

	dsSchema["instance_id"].Required = true
	dsSchema["instance_id"].Computed = false

	return &schema.Resource{
		ReadContext: dataSourceScalewayDocumentDBDatabaseRead,
		Schema:      dsSchema,
	}
}

func dataSourceScalewayDocumentDBDatabaseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	_, region, err := documentDBAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	instanceID := expandID(d.Get("instance_id").(string))
	databaseName := d.Get("name").(string)

	id := resourceScalewayDocumentDBDatabaseID(region, instanceID, databaseName)
	d.SetId(id)
	if err != nil {
		return diag.FromErr(err)
	}

	diags := resourceScalewayDocumentDBDatabaseRead(ctx, d, meta)
	if diags != nil {
		return append(diags, diag.Errorf("failed to read database state")...)
	}

	if d.Id() == "" {
		return diag.Errorf("database (%s) not found", databaseName)
	}

	return nil
}
