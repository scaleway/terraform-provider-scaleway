package scaleway

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
)

func resourceScalewayComputeInstanceVolume() *schema.Resource {
	return &schema.Resource{
		Create: resourceScalewayComputeInstanceVolumeCreate,
		Read:   resourceScalewayComputeInstanceVolumeRead,
		Update: resourceScalewayComputeInstanceVolumeUpdate,
		Delete: resourceScalewayComputeInstanceVolumeDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: func() (interface{}, error) {
					// TODO generate
					return "foo", nil
				},
				Description: "the name of the volume",
			},
			"size": {
				Type:        schema.TypeInt,
				Optional:    true,
				ForceNew:    true,
				Description: "the size of the volume in bytes", // TODO: human readable
			},
			// TODO handle snapshot, base_volume
			"server_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "the server associated with this volume",
			},
			"project_id": projectIDSchema(),
			"zone":       zoneSchema(),
		},
	}
}

func resourceScalewayComputeInstanceVolumeCreate(d *schema.ResourceData, m interface{}) error {

	meta := m.(*Meta)
	instanceAPI := instance.NewAPI(meta.scwClient)

	zone, err := getZone(d, meta)
	if err != nil {
		return err
	}

	var (
		volumeName = d.Get("name").(string)
		volumeSize = uint64(d.Get("size").(int))
		projectID  = d.Get("project_id").(string)
	)

	res, err := instanceAPI.CreateVolume(&instance.CreateVolumeRequest{
		Zone:         zone,
		Name:         volumeName,
		Size:         &volumeSize,
		VolumeType:   instance.VolumeTypeLSsd,
		Organization: projectID,
	})
	if err != nil {
		return fmt.Errorf("couldn't create volume: %s", err)
	}

	d.SetId(newZonedId(zone, res.Volume.ID))

	return resourceScalewayComputeInstanceVolumeRead(d, m)
}

func resourceScalewayComputeInstanceVolumeRead(d *schema.ResourceData, m interface{}) error {

	meta := m.(*Meta)
	instanceAPI := instance.NewAPI(meta.scwClient)

	zone, id, err := parseZonedID(d.Id())
	if err != nil {
		return err
	}

	res, err := instanceAPI.GetVolume(&instance.GetVolumeRequest{
		VolumeID: id,
		Zone:     zone,
	})
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("couldn't read volume: %v", err)
	}

	d.Set("name", res.Volume.Name)
	d.Set("size", res.Volume.Size)
	d.Set("project_id", res.Volume.Organization)
	d.Set("zone", string(zone))

	if res.Volume.Server != nil {
		d.Set("server_id", res.Volume.Server.ID)
	} else {
		d.Set("server_id", nil)
	}

	return nil
}

func resourceScalewayComputeInstanceVolumeUpdate(d *schema.ResourceData, m interface{}) error {

	meta := m.(*Meta)
	instanceAPI := instance.NewAPI(meta.scwClient)

	zone, id, err := parseZonedID(d.Id())
	if err != nil {
		return err
	}

	if d.HasChange("name") {

		newName := d.Get("name").(string)

		_, err = instanceAPI.UpdateVolume(&instance.UpdateVolumeRequest{
			VolumeID: id,
			Zone:     zone,
			Name:     &newName,
		})
		if err != nil {
			return fmt.Errorf("couldn't update volume: %s", err)
		}

	}

	return resourceScalewayComputeInstanceVolumeRead(d, m)
}

func resourceScalewayComputeInstanceVolumeDelete(d *schema.ResourceData, m interface{}) error {

	meta := m.(*Meta)
	instanceAPI := instance.NewAPI(meta.scwClient)

	zone, id, err := parseZonedID(d.Id())
	if err != nil {
		return err
	}

	err = instanceAPI.DeleteVolume(&instance.DeleteVolumeRequest{
		VolumeID: id,
		Zone:     zone,
	})

	if err != nil && !is404Error(err) {
		return fmt.Errorf("couldn't delete volume: %v", err)
	}

	return nil
}
