package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	documentdb "github.com/scaleway/scaleway-sdk-go/api/documentdb/v1beta1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func dataSourceScalewayDocumentDBInstance() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasource.SchemaFromResourceSchema(resourceScalewayDocumentDBInstance().Schema)

	datasource.AddOptionalFieldsToSchema(dsSchema, "name", "region", "project_id")

	dsSchema["instance_id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The ID of the instance",
		ConflictsWith: []string{"name"},
		ValidateFunc:  verify.IsUUIDorUUIDWithLocality(),
	}

	return &schema.Resource{
		ReadContext: dataSourceScalewayDocumentDBInstanceRead,
		Schema:      dsSchema,
	}
}

func dataSourceScalewayDocumentDBInstanceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, err := documentDBAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	instanceID, instanceIDExists := d.GetOk("instance_id")
	if !instanceIDExists {
		instanceName := d.Get("name").(string)
		res, err := api.ListInstances(&documentdb.ListInstancesRequest{
			Region:    region,
			Name:      types.ExpandStringPtr(instanceName),
			ProjectID: types.ExpandStringPtr(d.Get("project_id")),
		})
		if err != nil {
			return diag.FromErr(err)
		}

		foundRawInstance, err := findExact(
			res.Instances,
			func(s *documentdb.Instance) bool { return s.Name == instanceName },
			instanceName,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		instanceID = foundRawInstance.ID
	}

	regionID := datasource.NewRegionalID(instanceID, region)
	d.SetId(regionID)
	err = d.Set("instance_id", regionID)
	if err != nil {
		return diag.FromErr(err)
	}

	diags := resourceScalewayDocumentDBInstanceRead(ctx, d, m)
	if diags != nil {
		return append(diags, diag.Errorf("failed to read instance state")...)
	}

	if d.Id() == "" {
		return diag.Errorf("instance (%s) not found", regionID)
	}

	return nil
}
