package instance

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceVolume() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasource.SchemaFromResourceSchema(ResourceVolume().Schema)

	// Set 'Optional' schema elements
	datasource.AddOptionalFieldsToSchema(dsSchema, "name", "zone", "project_id")

	dsSchema["volume_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "The ID of the volume",
		ConflictsWith:    []string{"name"},
		ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
	}
	dsSchema["name"].ConflictsWith = []string{"volume_id"}

	return &schema.Resource{
		ReadContext: DataSourceInstanceVolumeRead,
		Schema:      dsSchema,
	}
}

func DataSourceInstanceVolumeRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	instanceAPI, zone, err := newAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	volumeID, ok := d.GetOk("volume_id")
	if !ok { // Get volumes by zone and name.
		volumeName := d.Get("name").(string)
		res, err := instanceAPI.ListVolumes(&instance.ListVolumesRequest{
			Zone:    zone,
			Name:    types.ExpandStringPtr(volumeName),
			Project: types.ExpandStringPtr(d.Get("project_id")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		foundVolume, err := datasource.FindExact(
			res.Volumes,
			func(s *instance.Volume) bool { return s.Name == volumeName },
			volumeName,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		volumeID = foundVolume.ID
	}

	zonedID := datasource.NewZonedID(volumeID, zone)
	d.SetId(zonedID)
	err = d.Set("volume_id", zonedID)
	if err != nil {
		return diag.FromErr(err)
	}
	return ResourceInstanceVolumeRead(ctx, d, m)
}
