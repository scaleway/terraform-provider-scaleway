package scaleway

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	namesgenerator "github.com/scaleway/scaleway-sdk-go/namegenerator"
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
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				Description:      "the name of the volume",
				DiffSuppressFunc: DiffSuppressFuncForRandomName,
			},
			"size_in_gb": {
				Type:        schema.TypeInt,
				Optional:    true,
				ForceNew:    true,
				Description: "the size of the volume in gigabye.",
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

	instanceAPI, zone, err := getInstanceAPIWithZone(d, m)
	if err != nil {
		return err
	}

	var (
		volumeName = d.Get("name").(string)
		volumeSize = uint64(d.Get("size_in_gb").(int))
		projectID  = d.Get("project_id").(string)
	)

	// Convert human readable volume size to int in bytes
	volumeSizeInBytes := volumeSize * gb

	// Generate name if not set
	if volumeName == "" {
		volumeName = namesgenerator.GetRandomName()
	}

	res, err := instanceAPI.CreateVolume(&instance.CreateVolumeRequest{
		Zone:         zone,
		Name:         volumeName,
		Size:         &volumeSizeInBytes,
		VolumeType:   instance.VolumeTypeLSSD,
		Organization: projectID,
	})
	if err != nil {
		return fmt.Errorf("couldn't create volume: %s", err)
	}

	d.SetId(newZonedId(zone, res.Volume.ID))

	return resourceScalewayComputeInstanceVolumeRead(d, m)
}

func resourceScalewayComputeInstanceVolumeRead(d *schema.ResourceData, m interface{}) error {

	instanceAPI, zone, id, err := getInstanceAPIWithZoneAndID(m, d.Id())
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
	d.Set("size_in_gb", uint64(res.Volume.Size/gb))
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

	instanceAPI, zone, id, err := getInstanceAPIWithZoneAndID(m, d.Id())
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

	instanceAPI, zone, id, err := getInstanceAPIWithZoneAndID(m, d.Id())
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
