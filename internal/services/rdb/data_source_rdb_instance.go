package rdb

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceInstance() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasource.SchemaFromResourceSchema(ResourceInstance().Schema)
	// Set 'Optional' schema elements
	datasource.AddOptionalFieldsToSchema(dsSchema, "name", "region", "project_id")

	dsSchema["name"].ConflictsWith = []string{"instance_id"}
	dsSchema["instance_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "The ID of the RDB instance",
		ConflictsWith:    []string{"name"},
		ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
	}

	return &schema.Resource{
		ReadContext: DataSourceRDBInstanceRead,
		Schema:      dsSchema,
	}
}

func DataSourceRDBInstanceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, err := newAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	instanceID, ok := d.GetOk("instance_id")
	if !ok { // Get instance by region and name.
		instanceName := d.Get("name").(string)

		res, err := api.ListInstances(&rdb.ListInstancesRequest{
			Region:    region,
			Name:      scw.StringPtr(instanceName),
			ProjectID: types.ExpandStringPtr(d.Get("project_id")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		foundInstance, err := datasource.FindExact(
			res.Instances,
			func(s *rdb.Instance) bool { return s.Name == instanceName },
			instanceName,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		instanceID = foundInstance.ID
	}

	regionalID := datasource.NewRegionalID(instanceID, region)
	d.SetId(regionalID)

	err = d.Set("instance_id", regionalID)
	if err != nil {
		return diag.FromErr(err)
	}

	return ResourceRdbInstanceRead(ctx, d, m)
}
