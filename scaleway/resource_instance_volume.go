package scaleway

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayInstanceVolume() *schema.Resource {
	return &schema.Resource{
		Create: resourceSalewayInstanceVolumeCreate,
		Read:   resourceSalewayInstanceVolumeRead,
		Update: resourceSalewayInstanceVolumeUpdate,
		Delete: resourceSalewayInstanceVolumeDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The name of the volume",
			},
			"type": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The volume type",
				ValidateFunc: validation.StringInSlice([]string{
					instance.VolumeTypeBSSD.String(),
					instance.VolumeTypeLSSD.String(),
				}, false),
			},
			"size_in_gb": {
				Type:          schema.TypeInt,
				Optional:      true,
				ForceNew:      true,
				Description:   "The size of the volume in gigabyte",
				ConflictsWith: []string{"from_snapshot_id", "from_volume_id"},
			},
			"from_volume_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				Description:   "Create a copy of an existing volume",
				ValidateFunc:  validationUUID(),
				ConflictsWith: []string{"from_snapshot_id", "size_in_gb"},
			},
			"from_snapshot_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				Description:   "Create a volume based on a image",
				ValidateFunc:  validationUUID(),
				ConflictsWith: []string{"from_volume_id", "size_in_gb"},
			},
			"server_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The server associated with this volume",
			},
			"organization_id": organizationIDSchema(),
			"zone":            zoneSchema(),
		},
	}
}

func resourceSalewayInstanceVolumeCreate(d *schema.ResourceData, m interface{}) error {
	instanceAPI, zone, err := getInstanceAPIWithZone(d, m)
	if err != nil {
		return err
	}

	createVolumeRequest := &instance.CreateVolumeRequest{
		Zone:         zone,
		Name:         d.Get("name").(string),
		VolumeType:   instance.VolumeType(d.Get("type").(string)),
		Organization: d.Get("organization_id").(string),
	}

	// Generate name if not set
	if createVolumeRequest.Name == "" {
		createVolumeRequest.Name = getRandomName("vol")
	}

	if size, ok := d.GetOk("size_in_gb"); ok {
		volumeSizeInBytes := scw.Size(uint64(size.(int)) * gb)
		createVolumeRequest.Size = &volumeSizeInBytes
	}

	if volumeID, ok := d.GetOk("from_volume_id"); ok {
		createVolumeRequest.BaseVolume = scw.StringPtr(expandID(volumeID))
	}

	if snapshotID, ok := d.GetOk("from_snapshot_id"); ok {
		createVolumeRequest.BaseSnapshot = scw.StringPtr(expandID(snapshotID))
	}

	res, err := instanceAPI.CreateVolume(createVolumeRequest)
	if err != nil {
		return fmt.Errorf("couldn't create volume: %s", err)
	}

	d.SetId(newZonedId(zone, res.Volume.ID))

	return resourceSalewayInstanceVolumeRead(d, m)
}

func resourceSalewayInstanceVolumeRead(d *schema.ResourceData, m interface{}) error {
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
	d.Set("organization_id", res.Volume.Organization)
	d.Set("zone", string(zone))

	if res.Volume.Server != nil {
		d.Set("server_id", res.Volume.Server.ID)
	} else {
		d.Set("server_id", nil)
	}

	return nil
}

func resourceSalewayInstanceVolumeUpdate(d *schema.ResourceData, m interface{}) error {
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

	return resourceSalewayInstanceVolumeRead(d, m)
}

func resourceSalewayInstanceVolumeDelete(d *schema.ResourceData, m interface{}) error {
	instanceAPI, zone, id, err := getInstanceAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return err
	}

	err = detachVolume(instanceAPI, zone, id)
	if err != nil {
		return err
	}

	deleteRequest := &instance.DeleteVolumeRequest{
		Zone:     zone,
		VolumeID: id,
	}

	err = instanceAPI.DeleteVolume(deleteRequest)
	return err
}
