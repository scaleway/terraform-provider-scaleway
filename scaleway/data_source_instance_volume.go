package scaleway

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	instance "github.com/scaleway/scaleway-sdk-go/api/instance/v1"
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
		Read:   dataSourceScalewayInstanceVolumeRead,
		Schema: dsSchema,
	}
}

func dataSourceScalewayInstanceVolumeRead(d *schema.ResourceData, m interface{}) error {
	meta := m.(*Meta)
	instanceApi, zone, err := instanceAPIWithZone(d, meta)
	if err != nil {
		return err
	}

	volumeID, ok := d.GetOk("volume_id")
	if !ok { // Get volumes by zone and name.
		res, err := instanceApi.ListVolumes(&instance.ListVolumesRequest{
			Zone: zone,
			Name: String(d.Get("name").(string)),
		})
		if err != nil {
			return err
		}
		if len(res.Volumes) == 0 {
			return fmt.Errorf("no volume found with the name %s", d.Get("name"))
		}
		if len(res.Volumes) > 1 {
			return fmt.Errorf("%d volumes found with the same name %s", len(res.Volumes), d.Get("name"))
		}
		volumeID = res.Volumes[0].ID
	}

	zonedID := datasourceNewZonedID(volumeID, zone)
	d.SetId(zonedID)
	err = d.Set("volume_id", zonedID)
	if err != nil {
		return err
	}
	return resourceScalewayInstanceVolumeRead(d, m)
}
