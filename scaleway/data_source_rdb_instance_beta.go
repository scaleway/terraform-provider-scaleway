package scaleway

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func dataSourceScalewayRDBInstance() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayRdbInstanceBeta().Schema)

	dsSchema["instance_id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The ID of the RDB instance",
		ConflictsWith: []string{"name"},
		ValidateFunc:  validationUUIDorUUIDWithLocality(),
	}

	return &schema.Resource{
		Read:   dataSourceScalewayRDBInstanceRead,
		Schema: dsSchema,
	}
}

func dataSourceScalewayRDBInstanceRead(d *schema.ResourceData, m interface{}) error {
	api, region, err := rdbAPIWithRegion(d, m)
	if err != nil {
		return err
	}

	instanceID, ok := d.GetOk("instance_id")
	if !ok { // Get instance by region and name.
		res, err := api.ListInstances(&rdb.ListInstancesRequest{
			Region: region,
			Name:   scw.StringPtr(d.Get("name").(string)),
		})
		if err != nil {
			return err
		}
		if len(res.Instances) == 0 {
			return fmt.Errorf("no instances found with the name %s", d.Get("name"))
		}
		if len(res.Instances) > 1 {
			return fmt.Errorf("%d instances found with the same name %s", len(res.Instances), d.Get("name"))
		}
		instanceID = res.Instances[0].ID
	}

	regionalID := datasourceNewRegionalizedID(instanceID, region)
	d.SetId(regionalID)
	err = d.Set("instance_id", regionalID)
	if err != nil {
		return err
	}
	return resourceScalewayRdbInstanceBetaRead(d, m)
}
