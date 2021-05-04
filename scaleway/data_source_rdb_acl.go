package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceScalewayRDBACL() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayRdbACL().Schema)

	dsSchema["instance_id"].Computed = false
	dsSchema["instance_id"].Required = true

	return &schema.Resource{
		ReadContext: dataSourceScalewayRDBACLRead,
		Schema:      dsSchema,
	}
}

func dataSourceScalewayRDBACLRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	_, region, err := rdbAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	instanceID, _ := d.GetOk("instance_id")

	_, _, err = parseLocalizedID(instanceID.(string))
	regionalID := instanceID
	if err != nil {
		regionalID = datasourceNewRegionalizedID(instanceID, region)
	}

	d.SetId(regionalID.(string))
	err = d.Set("instance_id", regionalID)
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceScalewayRdbACLRead(ctx, d, meta)
}
