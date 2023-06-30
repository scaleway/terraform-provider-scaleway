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
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayRdbInstance().Schema)
	// Set 'Optional' schema elements
	addOptionalFieldsToSchema(dsSchema, "name", "region")

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

func dataSourceScalewayRDBInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, err := rdbAPIWithRegion(d, meta)
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
		for _, instance := range res.Instances {
			if instance.Name == d.Get("name").(string) {
				if instanceID != "" {
					return diag.FromErr(fmt.Errorf("more than 1 instance found with the same name %s", d.Get("name")))
				}
				instanceID = instance.ID
			}
		}
		if instanceID == "" {
			return diag.FromErr(fmt.Errorf("no instance found with the name %s", d.Get("name")))
		}
	}

	regionalID := datasourceNewRegionalID(instanceID, region)
	d.SetId(regionalID)
	err = d.Set("instance_id", regionalID)
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceScalewayRdbInstanceRead(ctx, d, meta)
}
