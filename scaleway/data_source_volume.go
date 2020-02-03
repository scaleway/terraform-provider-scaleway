package scaleway

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	api "github.com/nicolai86/scaleway-sdk"
)

func dataSourceScalewayVolume() *schema.Resource {
	return &schema.Resource{
		DeprecationMessage: `This data source is deprecated and will be removed in the next major version.
		Please use scaleway_instance_volume instead.`,
		Read: dataSourceScalewayVolumeRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "the name of the volume",
			},
			"size_in_gb": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "the size of the volume in GB",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "the type of backing storage",
			},
			"server": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceScalewayVolumeRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*Meta).deprecatedClient

	name := d.Get("name").(string)

	volumes, err := client.GetVolumes()
	if err != nil {
		if serr, ok := err.(api.APIError); ok {
			log.Printf("[DEBUG] Error obtaining Volumes: %q\n", serr.APIMessage)
		}

		return fmt.Errorf("Error obtaining Volumes: %+v", err)
	}

	var volume *api.Volume
	for _, v := range *volumes {
		if v.Name == name {
			volume = &v
			break
		}
	}

	if volume == nil {
		return fmt.Errorf("Couldn't locate a Volume with the name %q!", name)
	}

	d.SetId(volume.Identifier)

	_ = d.Set("name", volume.Name)
	_ = d.Set("size_in_gb", int(uint64(volume.Size)/gb))
	_ = d.Set("type", volume.VolumeType)

	if volume.Server != nil {
		_ = d.Set("server", volume.Server.Identifier)
	} else {
		_ = d.Set("server", "")
	}

	return nil
}
