package iam

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
)

func DataSourceAPIKey() *schema.Resource {
	dsSchema := datasource.SchemaFromResourceSchema(ResourceAPIKey().Schema)

	dsSchema["access_key"].Required = true
	dsSchema["access_key"].Computed = false
	delete(dsSchema, "secret_key")

	return &schema.Resource{
		ReadContext: DataSourceIamAPIKeyRead,
		Schema:      dsSchema,
	}
}

func DataSourceIamAPIKeyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	accessKey := d.Get("access_key").(string)

	d.SetId(accessKey)

	diags := resourceIamAPIKeyRead(ctx, d, m)
	if diags != nil {
		return append(diags, diag.Errorf("failed to read iam api key state")...)
	}

	if d.Id() == "" {
		return diag.Errorf("iam api key (%s) not found", accessKey)
	}

	return nil
}
