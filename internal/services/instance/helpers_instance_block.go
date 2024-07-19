package instance

import (
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	block "github.com/scaleway/scaleway-sdk-go/api/block/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/api/marketplace/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

type BlockAndInstanceAPI struct {
	*instance.API
	blockAPI *block.API
}

type GetUnknownVolumeRequest struct {
	VolumeID string
	Zone     scw.Zone
}

type UnknownVolume struct {
	Zone     scw.Zone
	ID       string
	Name     string
	Size     *scw.Size
	ServerID *string
	Boot     *bool

	// Iops is set for Block volume only, use IsBlockVolume
	// Can be nil if not available in the Block API.
	Iops *uint32

	InstanceVolumeType instance.VolumeVolumeType
}

// VolumeTemplate returns a template to be used for servers requests.
func (volume *UnknownVolume) VolumeTemplate() *instance.VolumeServerTemplate {
	template := &instance.VolumeServerTemplate{}

	if volume.ID != "" {
		template.ID = &volume.ID
		if !volume.IsBlockVolume() {
			template.Name = &volume.Name
		}
	} else {
		template.VolumeType = volume.InstanceVolumeType
		template.Size = volume.Size
	}

	if volume.Boot != nil {
		template.Boot = volume.Boot
	}

	if volume.IsBlockVolume() {
		template.VolumeType = volume.InstanceVolumeType
	}

	return template
}

// IsLocal returns true if the volume is a local volume
func (volume *UnknownVolume) IsLocal() bool {
	return !volume.IsBlockVolume() && volume.InstanceVolumeType == instance.VolumeVolumeTypeLSSD
}

// IsBlockVolume is true if volume is managed by block API
func (volume *UnknownVolume) IsBlockVolume() bool {
	return volume.InstanceVolumeType == instance.VolumeVolumeTypeSbsVolume
}

// IsAttached returns true if the volume is attached to a server
func (volume *UnknownVolume) IsAttached() bool {
	return volume.ServerID != nil && *volume.ServerID != ""
}

func (api *BlockAndInstanceAPI) GetUnknownVolume(req *GetUnknownVolumeRequest, opts ...scw.RequestOption) (*UnknownVolume, error) {
	getVolumeResponse, err := api.API.GetVolume(&instance.GetVolumeRequest{
		Zone:     req.Zone,
		VolumeID: req.VolumeID,
	}, opts...)
	notFoundErr := &scw.ResourceNotFoundError{}
	if err != nil && !errors.As(err, &notFoundErr) {
		return nil, err
	}

	if getVolumeResponse != nil {
		vol := &UnknownVolume{
			Zone:               getVolumeResponse.Volume.Zone,
			ID:                 getVolumeResponse.Volume.ID,
			Name:               getVolumeResponse.Volume.Name,
			Size:               &getVolumeResponse.Volume.Size,
			InstanceVolumeType: getVolumeResponse.Volume.VolumeType,
		}
		if getVolumeResponse.Volume.Server != nil {
			vol.ServerID = &getVolumeResponse.Volume.Server.ID
		}

		return vol, nil
	}

	blockVolume, err := api.blockAPI.GetVolume(&block.GetVolumeRequest{
		Zone:     req.Zone,
		VolumeID: req.VolumeID,
	}, opts...)
	if err != nil {
		return nil, err
	}

	vol := &UnknownVolume{
		Zone:               blockVolume.Zone,
		ID:                 blockVolume.ID,
		Name:               blockVolume.Name,
		Size:               &blockVolume.Size,
		InstanceVolumeType: instance.VolumeVolumeTypeSbsVolume,
	}
	if blockVolume.Specs != nil {
		vol.Iops = blockVolume.Specs.PerfIops
	}
	for _, ref := range blockVolume.References {
		if ref.ProductResourceType == "instance_server" {
			vol.ServerID = &ref.ProductResourceID
		}
	}

	return vol, nil
}

// newAPIWithZone returns a new instance API and the zone for a Create request
func instanceAndBlockAPIWithZone(d *schema.ResourceData, m interface{}) (*BlockAndInstanceAPI, scw.Zone, error) {
	instanceAPI := instance.NewAPI(meta.ExtractScwClient(m))
	blockAPI := block.NewAPI(meta.ExtractScwClient(m))

	zone, err := meta.ExtractZone(d, m)
	if err != nil {
		return nil, "", err
	}

	return &BlockAndInstanceAPI{
		API:      instanceAPI,
		blockAPI: blockAPI,
	}, zone, nil
}

// NewAPIWithZoneAndID returns an instance API with zone and ID extracted from the state
func instanceAndBlockAPIWithZoneAndID(m interface{}, zonedID string) (*BlockAndInstanceAPI, scw.Zone, string, error) {
	instanceAPI := instance.NewAPI(meta.ExtractScwClient(m))
	blockAPI := block.NewAPI(meta.ExtractScwClient(m))

	zone, ID, err := zonal.ParseID(zonedID)
	if err != nil {
		return nil, "", "", err
	}

	return &BlockAndInstanceAPI{
		API:      instanceAPI,
		blockAPI: blockAPI,
	}, zone, ID, nil
}

func volumeTypeToMarketplaceFilter(volumeType any) marketplace.LocalImageType {
	if volumeType != nil && instance.VolumeVolumeType(volumeType.(string)) == instance.VolumeVolumeTypeSbsVolume {
		return marketplace.LocalImageTypeInstanceSbs
	} else {
		return marketplace.LocalImageTypeInstanceLocal
	}
}
