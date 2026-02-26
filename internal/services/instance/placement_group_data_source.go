package instance

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

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
	}
}

func DataSourcePlacementGroupRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, err := newAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	var pg *instance.PlacementGroup

	placementGroupID, placementGroupIDExists := d.GetOk("placement_group_id")
	if !placementGroupIDExists {
		name := d.Get("name").(string)

		res, err := api.ListPlacementGroups(&instance.ListPlacementGroupsRequest{
			Zone:    zone,
			Name:    types.ExpandStringPtr(name),
			Project: types.ExpandStringPtr(d.Get("project_id")),
		})
		if err != nil {
			return diag.FromErr(err)
		}

		for _, placementGroup := range res.PlacementGroups {
			if placementGroup.Name == name {
				if placementGroupID != "" {
					return diag.Errorf("more than 1 placement group found with the same name %s", name)
				}

				pg = placementGroup
				placementGroupID = placementGroup.ID
			}
		}

		if placementGroupID == "" {
			return diag.Errorf("no placement group found with the name %s", name)
		}
	} else {
		id, err := locality.ExtractUUID(placementGroupID.(string))
		if err != nil {
			return diag.FromErr(err)
		}

		res, err := api.GetPlacementGroup(&instance.GetPlacementGroupRequest{
			Zone:             zone,
			PlacementGroupID: id,
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		pg = res.PlacementGroup
	}

	zoneID := datasource.NewZonedID(placementGroupID, zone)
	d.SetId(zoneID)

	err = d.Set("placement_group_id", zoneID)
	if err != nil {
		return diag.FromErr(err)
	}

	return setPlacementGroupState(d, pg)
}
