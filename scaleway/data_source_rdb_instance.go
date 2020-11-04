package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func dataSourceScalewayRDBInstance() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayRdbInstanceBeta().Schema)
	// Set 'Optional' schema elements
	addOptionalFieldsToSchema(dsSchema, "name")

	dsSchema["name"].ConflictsWith = []string{"instance_id"}
	dsSchema["instance_id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The ID of the RDB instance",
		ConflictsWith: []string{"name"},
		ValidateFunc:  validationUUIDorUUIDWithLocality(),
	}

	return &schema.Resource{
		ReadContext: dataSourceScalewayRDBInstanceRead,
		Schema:      dsSchema,
	}
}

func dataSourceScalewayRDBInstanceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, err := rdbAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	instanceID, ok := d.GetOk("instance_id")
	if !ok { // Get instance by region and name.
		res, err := api.ListInstances(&rdb.ListInstancesRequest{
			Region: region,
			Name:   scw.StringPtr(d.Get("name").(string)),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
		if len(res.Instances) == 0 {
			return diag.FromErr(fmt.Errorf("no instances found with the name %s", d.Get("name")))
		}
		if len(res.Instances) > 1 {
			return diag.FromErr(fmt.Errorf("%d instances found with the same name %s", len(res.Instances), d.Get("name")))
		}
		instanceID = res.Instances[0].ID
	}

	regionalID := datasourceNewRegionalizedID(instanceID, region)
	d.SetId(regionalID)
	err = d.Set("instance_id", regionalID)
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceScalewayRdbInstanceBetaRead(ctx, d, m)
}
