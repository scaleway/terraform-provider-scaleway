package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
)

func dataSourceScalewayRDBACL() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayRdbACL().Schema)

	dsSchema["instance_id"].Computed = false
	dsSchema["instance_id"].Required = true

	// Set 'Optional' schema elements
	addOptionalFieldsToSchema(dsSchema, "region")
	return &schema.Resource{
		ReadContext: dataSourceScalewayRDBACLRead,
		Schema:      dsSchema,
	}
}

func dataSourceScalewayRDBACLRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	_, region, err := rdbAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	instanceID, _ := d.GetOk("instance_id")

	_, _, err = locality.ParseLocalizedID(instanceID.(string))
	regionalID := instanceID
	if err != nil {
		regionalID = datasourceNewRegionalID(instanceID, region)
	}

	d.SetId(regionalID.(string))
	err = d.Set("instance_id", regionalID)
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceScalewayRdbACLRead(ctx, d, m)
}
