package rdb

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
)

func DataSourceACL() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasource.SchemaFromResourceSchema(ResourceACL().Schema)

	dsSchema["instance_id"].Computed = false
	dsSchema["instance_id"].Required = true

	// Set 'Optional' schema elements
	datasource.AddOptionalFieldsToSchema(dsSchema, "region")

	return &schema.Resource{
		ReadContext: DataSourceRDBACLRead,
		Schema:      dsSchema,
	}
}

func DataSourceRDBACLRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	_, region, err := newAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	instanceID, _ := d.GetOk("instance_id")

	_, _, err = locality.ParseLocalizedID(instanceID.(string))
	regionalID := instanceID
	if err != nil {
		regionalID = datasource.NewRegionalID(instanceID, region)
	}

	d.SetId(regionalID.(string))
	err = d.Set("instance_id", regionalID)
	if err != nil {
		return diag.FromErr(err)
	}

	return ResourceRdbACLRead(ctx, d, m)
}
