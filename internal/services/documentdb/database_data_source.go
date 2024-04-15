package documentdb

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
)

func DataSourceDatabase() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasource.SchemaFromResourceSchema(ResourceDatabase().Schema)

	datasource.AddOptionalFieldsToSchema(dsSchema, "name", "region")

	dsSchema["instance_id"].Required = true
	dsSchema["instance_id"].Computed = false

	return &schema.Resource{
		ReadContext: DataSourceDocumentDBDatabaseRead,
		Schema:      dsSchema,
	}
}

func DataSourceDocumentDBDatabaseRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	_, region, err := documentDBAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	instanceID := locality.ExpandID(d.Get("instance_id").(string))
	databaseName := d.Get("name").(string)

	id := resourceDocumentDBDatabaseID(region, instanceID, databaseName)
	d.SetId(id)
	if err != nil {
		return diag.FromErr(err)
	}

	diags := resourceDocumentDBDatabaseRead(ctx, d, m)
	if diags != nil {
		return append(diags, diag.Errorf("failed to read database state")...)
	}

	if d.Id() == "" {
		return diag.Errorf("database (%s) not found", databaseName)
	}

	return nil
}
