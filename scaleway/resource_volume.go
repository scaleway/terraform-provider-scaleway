package scaleway

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	api "github.com/nicolai86/scaleway-sdk"
)

const gb uint64 = 1000 * 1000 * 1000

func resourceScalewayVolume() *schema.Resource {
	return &schema.Resource{
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
					if value < 1 || value > 150 {
						errors = append(errors, fmt.Errorf("%q be more than 1 and less than 150", k))
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
	scaleway := m.(*Client).scaleway

	mu.Lock()
	defer mu.Unlock()

	size := uint64(d.Get("size_in_gb").(int)) * gb
	req := api.VolumeDefinition{
		Name:         d.Get("name").(string),
		Size:         size,
		Type:         d.Get("type").(string),
		Organization: scaleway.Organization,
	}
	var (
		volumeID string
		err      error
	)
	if err := retry(func() error {
		volumeID, err = scaleway.PostVolume(req)
		return err
	}); err != nil {
		return fmt.Errorf("Error Creating volume: %q", err)
	}
	d.SetId(volumeID)
	return resourceScalewayVolumeRead(d, m)
}

func resourceScalewayVolumeRead(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Client).scaleway
	var (
		volume *api.Volume
		err    error
	)
	if err := retry(func() error {
		volume, err = scaleway.GetVolume(d.Id())
		return err
	}); err != nil {
		if serr, ok := err.(api.APIError); ok {
			log.Printf("[DEBUG] Error reading volume: %q\n", serr.APIMessage)

			if serr.StatusCode == 404 {
				d.SetId("")
				return nil
			}
		}

		return err
	}
	d.Set("name", volume.Name)
	d.Set("size_in_gb", uint64(volume.Size)/gb)
	d.Set("type", volume.VolumeType)
	d.Set("server", "")
	if volume.Server != nil {
		d.Set("server", volume.Server.Identifier)
	}
	return nil
}

func resourceScalewayVolumeUpdate(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Client).scaleway

	mu.Lock()
	defer mu.Unlock()

	var req api.VolumePutDefinition
	if d.HasChange("name") {
		req.Name = String(d.Get("name").(string))
	}

	if d.HasChange("size_in_gb") {
		size := uint64(d.Get("size_in_gb").(int)) * gb
		req.Size = &size
	}

	retry(func() error {
		return scaleway.PutVolume(d.Id(), req)
	})
	return resourceScalewayVolumeRead(d, m)
}

func resourceScalewayVolumeDelete(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Client).scaleway

	mu.Lock()
	defer mu.Unlock()

	if err := retry(func() error {
		return scaleway.DeleteVolume(d.Id())
	}); err != nil {
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
