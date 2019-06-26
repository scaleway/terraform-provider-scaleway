package scaleway

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/utils"
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
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "the name of the volume",
			},
			"type": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     instance.VolumeTypeLSSD.String(),
				Description: "the volume type.",
			},
			"size_in_gb": {
				Type:          schema.TypeInt,
				Optional:      true,
				ForceNew:      true,
				Description:   "the size of the volume in gigabyte.",
				ConflictsWith: []string{"from_image_id", "from_volume_id"},
			},
			"from_volume_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				Description:   "create a copy of an existing volume.",
				ValidateFunc:  validationUUID(),
				ConflictsWith: []string{"from_image_id", "size_in_gb"},
			},
			"from_image_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				Description:   "create a volume based on a image.",
				ValidateFunc:  validationUUID(),
				ConflictsWith: []string{"from_volume_id", "size_in_gb"},
			},
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

	createVolumeRequest := &instance.CreateVolumeRequest{
		Zone:         zone,
		Name:         d.Get("name").(string),
		VolumeType:   instance.VolumeType(d.Get("type").(string)),
		Organization: d.Get("project_id").(string),
	}

	// Generate name if not set
	if createVolumeRequest.Name == "" {
		createVolumeRequest.Name = getRandomName("vol")
	}

	if size, ok := d.GetOk("size_in_gb"); ok {
		volumeSizeInBytes := uint64(size.(int)) * gb
		createVolumeRequest.Size = &volumeSizeInBytes
	}

	if volumeID, ok := d.GetOk("from_volume_id"); ok {
		createVolumeRequest.BaseVolume = utils.String(expandID(volumeID))
	}

	if imageID, ok := d.GetOk("from_image_id"); ok {
		createVolumeRequest.BaseSnapshot = utils.String(expandID(imageID))
	}

	res, err := instanceAPI.CreateVolume(createVolumeRequest)
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
	d.Set("project_id", res.Volume.Organization)
	d.Set("zone", string(zone))

	if res.Volume.Server != nil {
		d.Set("server_id", res.Volume.Server.ID)
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

	getVolumeResponse, err := instanceAPI.GetVolume(&instance.GetVolumeRequest{
		Zone:     zone,
		VolumeID: id,
	})
	if err != nil {
		return err
	}

	if getVolumeResponse.Volume.Server != nil {
		instanceAPI.ServerActionAndWait(&instance.ServerActionAndWaitRequest{
			Zone:     zone,
			ServerID: getVolumeResponse.Volume.Server.ID,
			Action:   instance.ServerActionPoweroff,
			Timeout:  ServerWaitForTimeout,
		})
		// ignore errors
		_, err := instanceAPI.DetachVolume(&instance.DetachVolumeRequest{
			Zone:     zone,
			VolumeID: id,
		})
		if err != nil && !is404Error(err) {
			return err
		}
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
