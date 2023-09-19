package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	document_db "github.com/scaleway/scaleway-sdk-go/api/documentdb/v1beta1"
)

func dataSourceScalewayDocumentDBInstance() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayDocumentDBInstance().Schema)

	addOptionalFieldsToSchema(dsSchema, "name", "region")

	dsSchema["instance_id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The ID of the instance",
		ConflictsWith: []string{"name"},
		ValidateFunc:  validationUUIDorUUIDWithLocality(),
	}

	return &schema.Resource{
		ReadContext: dataSourceScalewayDocumentDBInstanceRead,
		Schema:      dsSchema,
	}
}

func dataSourceScalewayDocumentDBInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, err := documentDBAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	instanceID, instanceIDExists := d.GetOk("instance_id")
	if !instanceIDExists {
		res, err := api.ListInstances(&document_db.ListInstancesRequest{
			Region:    region,
			Name:      expandStringPtr(d.Get("name")),
			ProjectID: expandStringPtr(d.Get("project_id")),
		})
		if err != nil {
			return diag.FromErr(err)
		}
		for _, instance := range res.Instances {
			if instance.Name == d.Get("name").(string) {
				if instanceID != "" {
					return diag.Errorf("more than 1 instance found with the same name %s", d.Get("name"))
				}
				instanceID = instance.ID
			}
		}
		if instanceID == "" {
			return diag.Errorf("no instance found with the name %s", d.Get("name"))
		}
	}

	regionID := datasourceNewRegionalID(instanceID, region)
	d.SetId(regionID)
	err = d.Set("instance_id", regionID)
	if err != nil {
		return diag.FromErr(err)
	}

	diags := resourceScalewayDocumentDBInstanceRead(ctx, d, meta)
	if diags != nil {
		return append(diags, diag.Errorf("failed to read instance state")...)
	}

	if d.Id() == "" {
		return diag.Errorf("instance (%s) not found", regionID)
	}

	return nil
}
