package scaleway

import (
	"fmt"

	"github.com/dustin/go-humanize"
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
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "the name of the volume",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if new == "" {
						return true
					}
					return old == new
				},
			},
			"size": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "the size of the volume in humab readable format (e.g. 20GB)", // TODO: human readable
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					_, err := humanize.ParseBytes(val.(string))
					if err != nil {
						errs = append(errs, fmt.Errorf("couldn't parse volume size: %s", err))
					}
					return
				},
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if old == new {
						return true
					}

					oldInBytes, err := humanize.ParseBytes(old)
					if err != nil {
						return false
					}
					newInBytes, err := humanize.ParseBytes(new)
					if err != nil {
						return false
					}

					return newInBytes == oldInBytes
				},
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
		volumeSize = d.Get("size").(string)
		projectID  = d.Get("project_id").(string)
	)

	// Convert human readable volume size to int in bytes
	volumeSizeInBytes, err := humanize.ParseBytes(volumeSize)
	if err != nil {
		return fmt.Errorf("couldn't parse volume size: %s", err)
	}

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
