package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func dataSourceScalewayInstanceVolume() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayInstanceVolume().Schema)

	// Set 'Optional' schema elements
	addOptionalFieldsToSchema(dsSchema, "name", "zone")

	dsSchema["volume_id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The ID of the volume",
		ConflictsWith: []string{"name"},
		ValidateFunc:  validationUUIDorUUIDWithLocality(),
	}
	dsSchema["name"].ConflictsWith = []string{"volume_id"}

	return &schema.Resource{
		ReadContext: dataSourceScalewayInstanceVolumeRead,
		Schema:      dsSchema,
	}
}

func dataSourceScalewayInstanceVolumeRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := m.(*Meta)
	instanceAPI, zone, err := instanceAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	volumeID, ok := d.GetOk("volume_id")
	if !ok { // Get volumes by zone and name.
		res, err := instanceAPI.ListVolumes(&instance.ListVolumesRequest{
			Zone:    zone,
			Name:    expandStringPtr(d.Get("name")),
			Project: expandStringPtr(d.Get("project_id")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
		for _, volume := range res.Volumes {
			if volume.Name == d.Get("name").(string) {
				if volumeID != "" {
					return diag.FromErr(fmt.Errorf("more than 1 volume found with the same name %s", d.Get("name")))
				}
				volumeID = volume.ID
			}
		}
		if volumeID == "" {
			return diag.FromErr(fmt.Errorf("no volume found with the name %s", d.Get("name")))
		}
	}

	zonedID := datasourceNewZonedID(volumeID, zone)
	d.SetId(zonedID)
	err = d.Set("volume_id", zonedID)
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceScalewayInstanceVolumeRead(ctx, d, m)
}
