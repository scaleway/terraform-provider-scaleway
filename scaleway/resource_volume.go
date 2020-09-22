package scaleway

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	api "github.com/nicolai86/scaleway-sdk"
)

func resourceScalewayVolume() *schema.Resource {
	return &schema.Resource{
		DeprecationMessage: `This resource is deprecated and will be removed in the next major version.
 Please use scaleway_instance_volume instead.`,

		Create: resourceScalewayVolumeCreate,
		Read:   resourceScalewayVolumeRead,
		Update: resourceScalewayVolumeUpdate,
		Delete: resourceScalewayVolumeDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "the name of the volume",
			},
			"size_in_gb": {
				Type:     schema.TypeInt,
				Required: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(int)
					if value < 1 || value > 300 {
						errors = append(errors, fmt.Errorf("%q be more than 1 and less than 300", k))
					}
					return
				},
				Description: "the size of the volume in GB",
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateVolumeType,
				Description:  "the type of backing storage",
			},
			"server": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "the server the volume is attached to",
			},
		},
	}
}

func resourceScalewayVolumeCreate(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Meta).deprecatedClient

	size := uint64(d.Get("size_in_gb").(int)) * gb
	req := api.VolumeDefinition{
		Name:         d.Get("name").(string),
		Size:         size,
		Type:         d.Get("type").(string),
		Organization: scaleway.Organization,
	}
	v, err := scaleway.CreateVolume(req)
	if err != nil {
		return fmt.Errorf("Error Creating volume: %q", err)
	}
	d.SetId(v.Identifier)
	return resourceScalewayVolumeRead(d, m)
}

func resourceScalewayVolumeRead(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Meta).deprecatedClient
	volume, err := scaleway.GetVolume(d.Id())
	if err != nil {
		if serr, ok := err.(api.APIError); ok {
			log.Printf("[DEBUG] Error reading volume: %q\n", serr.APIMessage)

			if serr.StatusCode == 404 {
				d.SetId("")
				return nil
			}
		}

		return err
	}
	_ = d.Set("name", volume.Name)
	_ = d.Set("size_in_gb", uint64(volume.Size)/gb)
	_ = d.Set("type", volume.VolumeType)
	_ = d.Set("server", "")
	if volume.Server != nil {
		_ = d.Set("server", volume.Server.Identifier)
	}
	return nil
}

func resourceScalewayVolumeUpdate(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Meta).deprecatedClient

	var req api.VolumePutDefinition
	if d.HasChange("name") {
		req.Name = String(d.Get("name").(string))
	}

	if d.HasChange("size_in_gb") {
		size := uint64(d.Get("size_in_gb").(int)) * gb
		req.Size = &size
	}

	_, _ = scaleway.UpdateVolume(d.Id(), req)
	return resourceScalewayVolumeRead(d, m)
}

func resourceScalewayVolumeDelete(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Meta).deprecatedClient

	err := scaleway.DeleteVolume(d.Id())
	if err != nil {
		if serr, ok := err.(api.APIError); ok {
			if serr.StatusCode == 404 {
				d.SetId("")
				return nil
			}
		}
		return err
	}
	d.SetId("")
	return nil
}
