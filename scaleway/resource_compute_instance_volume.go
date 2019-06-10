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

func resourceScalewayComputeInstanceVolumeCreate(d *schema.ResourceData, m interface{}) error {

	scwClient := m.(*Meta).scwClient
	instanceAPI := instance.NewAPI(scwClient)

	// TODO add zone to resource config?

	/*
		defaultZone, ok := scwClient.GetDefaultZone()
		if !ok {
			return fmt.Errorf("couldn't create volume: default zone not set")
		}
	*/

	var (
		//zone       = defaultZone
		volumeName = d.Get("name").(string)
		volumeSize = uint64(d.Get("size_in_gb").(int)) * gbToBytesFactor
		volumeType = d.Get("type").(string)
		//volumeServer = d.Get("server").(string)
	)

	createVolumeResponse, err := instanceAPI.CreateVolume(&instance.CreateVolumeRequest{
		//Zone:       zone,
		Name:       volumeName,
		Size:       &volumeSize,
		VolumeType: instance.VolumeType(volumeType),
	})
	if err != nil {
		return fmt.Errorf("couldn't create volume: %s", err)
	}

	// attach to server

	d.SetId(createVolumeResponse.Volume.ID)

	return resourceScalewayComputeInstanceVolumeRead(d, m)
}

func resourceScalewayComputeInstanceVolumeRead(d *schema.ResourceData, m interface{}) error {

	scwClient := m.(*Meta).scwClient
	instanceAPI := instance.NewAPI(scwClient)

	getVolumeResponse, err := instanceAPI.GetVolume(&instance.GetVolumeRequest{
		VolumeID: d.Id(),
	})
	if err != nil {
		return fmt.Errorf("couldn't read volume: %s", err)
	}
	if getVolumeResponse.Volume == nil {
		return fmt.Errorf("couldn't read volume: received empty Volume in response")
	}

	d.Set("name", getVolumeResponse.Volume.Name)
	d.Set("size_in_gb", getVolumeResponse.Volume.Size/gbToBytesFactor)
	d.Set("type", getVolumeResponse.Volume.VolumeType)
	d.Set("server", getVolumeResponse.Volume.Server)

	return nil
}

func resourceScalewayComputeInstanceVolumeUpdate(d *schema.ResourceData, m interface{}) error {

	if d.HasChange("name") || d.HasChange("size_in_gb") {

		scwClient := m.(*Meta).scwClient
		instanceAPI := instance.NewAPI(scwClient)

		getVolumeResponse, err := instanceAPI.GetVolume(&instance.GetVolumeRequest{
			VolumeID: d.Id(),
		})
		if err != nil {
			return fmt.Errorf("couldn't update volume: %s", err)
		}

		_, err = instanceAPI.SetVolume(&instance.SetVolumeRequest{})
		if err != nil {
			return fmt.Errorf("couldn't update volume: %s", err)
		}

	}

	return resourceScalewayComputeInstanceVolumeRead(d, m)
}

func resourceScalewayComputeInstanceVolumeDelete(d *schema.ResourceData, m interface{}) error {

	scwClient := m.(*Meta).scwClient
	instanceAPI := instance.NewAPI(scwClient)

	err := instanceAPI.DeleteVolume(&instance.DeleteVolumeRequest{
		VolumeID: d.Id(),
	})

	if err != nil {
		return fmt.Errorf("couldn't delete volume: %s", err)
	}

	return nil
}

const gbToBytesFactor uint64 = 1000000000
