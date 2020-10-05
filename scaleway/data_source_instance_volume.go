package scaleway

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	instance "github.com/scaleway/scaleway-sdk-go/api/instance/v1"
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
		Read:   dataSourceScalewayInstanceVolumeRead,
		Schema: dsSchema,
	}
}

func dataSourceScalewayInstanceVolumeRead(d *schema.ResourceData, m interface{}) error {
	meta := m.(*Meta)
	instanceAPI, zone, err := instanceAPIWithZone(d, meta)
	if err != nil {
		return err
	}

	volumeID, ok := d.GetOk("volume_id")
	if !ok { // Get volumes by zone and name.
		res, err := instanceAPI.ListVolumes(&instance.ListVolumesRequest{
			Zone: zone,
			Name: scw.StringPtr(d.Get("name").(string)),
		})
		if err != nil {
			return err
		}
		for _, volume := range res.Volumes {
			if volume.Name == d.Get("name").(string) {
				if volumeID != "" {
					return fmt.Errorf("more than 1 volume found with the same name %s", d.Get("name"))
				}
				volumeID = volume.ID
			}
		}
		if volumeID == "" {
			return fmt.Errorf("no volume found with the name %s", d.Get("name"))
		}
	}

	zonedID := datasourceNewZonedID(volumeID, zone)
	d.SetId(zonedID)
	err = d.Set("volume_id", zonedID)
	if err != nil {
		return err
	}
	return resourceScalewayInstanceVolumeRead(d, m)
}
