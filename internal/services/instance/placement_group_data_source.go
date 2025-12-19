package instance

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

//go:embed descriptions/placement_group_datasource.md
var placementGroupDataSourceDescription string

func DataSourcePlacementGroup() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasource.SchemaFromResourceSchema(ResourcePlacementGroup().SchemaFunc())

	datasource.AddOptionalFieldsToSchema(dsSchema, "name", "zone")

	dsSchema["placement_group_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "The ID of the placementgroup",
		ConflictsWith:    []string{"name"},
		ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
	}
	dsSchema["project_id"].Optional = true

	return &schema.Resource{
		ReadContext: DataSourcePlacementGroupRead,
		Schema:      dsSchema,
		Description: placementGroupDataSourceDescription,
	}
}

func DataSourcePlacementGroupRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, err := newAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	placementGroupID, placementGroupIDExists := d.GetOk("placement_group_id")
	if !placementGroupIDExists {
		res, err := api.ListPlacementGroups(&instance.ListPlacementGroupsRequest{
			Zone:    zone,
			Name:    types.ExpandStringPtr(d.Get("name")),
			Project: types.ExpandStringPtr(d.Get("project_id")),
		})
		if err != nil {
			return diag.FromErr(err)
		}

		for _, placementGroup := range res.PlacementGroups {
			if placementGroup.Name == d.Get("name").(string) {
				if placementGroupID != "" {
					return diag.Errorf("more than 1 placement group found with the same name %s", d.Get("name"))
				}

				placementGroupID = placementGroup.ID
			}
		}

		if placementGroupID == "" {
			return diag.Errorf("no placementgroup found with the name %s", d.Get("name"))
		}
	}

	zoneID := datasource.NewZonedID(placementGroupID, zone)
	d.SetId(zoneID)

	err = d.Set("placement_group_id", zoneID)
	if err != nil {
		return diag.FromErr(err)
	}

	diags := ResourceInstancePlacementGroupRead(ctx, d, m)
	if diags != nil {
		return append(diags, diag.Errorf("failed to read placement group state")...)
	}

	if d.Id() == "" {
		return diag.Errorf("placement group (%s) not found", zoneID)
	}

	return nil
}
