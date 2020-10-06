package scaleway

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayInstanceVolume() *schema.Resource {
	return &schema.Resource{
		Create: resourceScalewayInstanceVolumeCreate,
		Read:   resourceScalewayInstanceVolumeRead,
		Update: resourceScalewayInstanceVolumeUpdate,
		Delete: resourceScalewayInstanceVolumeDelete,
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
					instance.VolumeVolumeTypeBSSD.String(),
					instance.VolumeVolumeTypeLSSD.String(),
				}, false),
			},
			"size_in_gb": {
				Type:          schema.TypeInt,
				Computed:      true,
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

func resourceScalewayInstanceVolumeCreate(d *schema.ResourceData, m interface{}) error {
	instanceAPI, zone, err := instanceAPIWithZone(d, m)
	if err != nil {
		return err
	}

	createVolumeRequest := &instance.CreateVolumeRequest{
		Zone:         zone,
		Name:         expandOrGenerateString(d.Get("name"), "vol"),
		VolumeType:   instance.VolumeVolumeType(d.Get("type").(string)),
		Organization: expandStringPtr(d.Get("organization_id")),
	}

	if size, ok := d.GetOk("size_in_gb"); ok {
		volumeSizeInBytes := scw.Size(uint64(size.(int)) * gb)
		createVolumeRequest.Size = &volumeSizeInBytes
	}

	if volumeID, ok := d.GetOk("from_volume_id"); ok {
		createVolumeRequest.BaseVolume = expandStringPtr(expandID(volumeID))
	}

	if snapshotID, ok := d.GetOk("from_snapshot_id"); ok {
		createVolumeRequest.BaseSnapshot = expandStringPtr(expandID(snapshotID))
	}

	res, err := instanceAPI.CreateVolume(createVolumeRequest)
	if err != nil {
		return fmt.Errorf("couldn't create volume: %s", err)
	}

	d.SetId(newZonedIDString(zone, res.Volume.ID))

	return resourceScalewayInstanceVolumeRead(d, m)
}

func resourceScalewayInstanceVolumeRead(d *schema.ResourceData, m interface{}) error {
	instanceAPI, zone, id, err := instanceAPIWithZoneAndID(m, d.Id())
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

	_ = d.Set("name", res.Volume.Name)
	_ = d.Set("organization_id", res.Volume.Organization)
	_ = d.Set("zone", string(zone))
	_ = d.Set("type", res.Volume.VolumeType.String())
	_ = d.Set("size_in_gb", uint64(res.Volume.Size/scw.GB))

	if res.Volume.Server != nil {
		_ = d.Set("server_id", res.Volume.Server.ID)
	} else {
		_ = d.Set("server_id", nil)
	}

	return nil
}

func resourceScalewayInstanceVolumeUpdate(d *schema.ResourceData, m interface{}) error {
	instanceAPI, zone, id, err := instanceAPIWithZoneAndID(m, d.Id())
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

	return resourceScalewayInstanceVolumeRead(d, m)
}

func resourceScalewayInstanceVolumeDelete(d *schema.ResourceData, m interface{}) error {
	instanceAPI, zone, id, err := instanceAPIWithZoneAndID(m, d.Id())
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
