package instancehelpers

import (
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	block "github.com/scaleway/scaleway-sdk-go/api/block/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

// InstanceAndBlockAPIWithZone returns a new instance API and the zone for a Create request
func InstanceAndBlockAPIWithZone(d *schema.ResourceData, m any) (*BlockAndInstanceAPI, scw.Zone, error) {
	zone, err := meta.ExtractZone(d, m)
	if err != nil {
		return nil, "", err
	}

	return NewBlockAndInstanceAPI(meta.ExtractScwClient(m)), zone, nil
}

// InstanceAndBlockAPIWithZoneAndID returns an instance API with zone and ID extracted from the state
func InstanceAndBlockAPIWithZoneAndID(m any, zonedID string) (*BlockAndInstanceAPI, scw.Zone, string, error) {
	zone, ID, err := zonal.ParseID(zonedID)
	if err != nil {
		return nil, "", "", err
	}

	return NewBlockAndInstanceAPI(meta.ExtractScwClient(m)), zone, ID, nil
}

type BlockAndInstanceAPI struct {
	*instance.API
	BlockAPI *block.API
}

func NewBlockAndInstanceAPI(client *scw.Client) *BlockAndInstanceAPI {
	instanceAPI := instance.NewAPI(client)
	blockAPI := block.NewAPI(client)

	return &BlockAndInstanceAPI{
		API:      instanceAPI,
		BlockAPI: blockAPI,
	}
}

type GetUnknownVolumeRequest struct {
	VolumeID string
	Zone     scw.Zone
}

type ResizeUnknownVolumeRequest struct {
	Size     *scw.Size
	VolumeID string
	Zone     scw.Zone
}

type DeleteUnknownVolumeRequest struct {
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

type UnknownSnapshot struct {
	Zone       scw.Zone
	ID         string
	Name       string
	VolumeType instance.VolumeVolumeType
}

func (api *BlockAndInstanceAPI) GetUnknownVolume(req *GetUnknownVolumeRequest, opts ...scw.RequestOption) (*UnknownVolume, error) {
	getVolumeResponse, err := api.GetVolume(&instance.GetVolumeRequest{
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

	blockVolume, err := api.BlockAPI.GetVolume(&block.GetVolumeRequest{
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

func (api *BlockAndInstanceAPI) ResizeUnknownVolume(req *ResizeUnknownVolumeRequest, opts ...scw.RequestOption) error {
	unknownVolume, err := api.GetUnknownVolume(&GetUnknownVolumeRequest{
		VolumeID: req.VolumeID,
		Zone:     req.Zone,
	}, opts...)
	if err != nil {
		return err
	}

	if unknownVolume.IsBlockVolume() {
		_, err = api.BlockAPI.UpdateVolume(&block.UpdateVolumeRequest{
			Zone:     req.Zone,
			VolumeID: req.VolumeID,
			Size:     req.Size,
		}, opts...)
	} else {
		_, err = api.UpdateVolume(&instance.UpdateVolumeRequest{
			Zone:     req.Zone,
			VolumeID: req.VolumeID,
			Size:     req.Size,
		}, opts...)
	}

	return err
}

func (api *BlockAndInstanceAPI) DeleteUnknownVolume(req *DeleteUnknownVolumeRequest, opts ...scw.RequestOption) error {
	unknownVolume, err := api.GetUnknownVolume(&GetUnknownVolumeRequest{
		VolumeID: req.VolumeID,
		Zone:     req.Zone,
	}, opts...)
	if err != nil {
		return err
	}

	if unknownVolume.IsBlockVolume() {
		targetStatus := block.VolumeStatusAvailable

		_, err = api.BlockAPI.WaitForVolume(&block.WaitForVolumeRequest{
			Zone:           req.Zone,
			VolumeID:       req.VolumeID,
			TerminalStatus: &targetStatus,
		}, opts...)
		if err != nil {
			return err
		}

		err = api.BlockAPI.DeleteVolume(&block.DeleteVolumeRequest{
			Zone:     req.Zone,
			VolumeID: req.VolumeID,
		}, opts...)
	} else {
		err = api.DeleteVolume(&instance.DeleteVolumeRequest{
			Zone:     req.Zone,
			VolumeID: req.VolumeID,
		}, opts...)
	}

	return err
}

type GetUnknownSnapshotRequest struct {
	Zone       scw.Zone
	SnapshotID string
}

func (api *BlockAndInstanceAPI) GetUnknownSnapshot(req *GetUnknownSnapshotRequest, opts ...scw.RequestOption) (*UnknownSnapshot, error) {
	getSnapshotResponse, err := api.GetSnapshot(&instance.GetSnapshotRequest{
		Zone:       req.Zone,
		SnapshotID: req.SnapshotID,
	}, opts...)
	notFoundErr := &scw.ResourceNotFoundError{}

	if err != nil && !errors.As(err, &notFoundErr) {
		return nil, err
	}

	if getSnapshotResponse != nil {
		snap := &UnknownSnapshot{
			Zone:       getSnapshotResponse.Snapshot.Zone,
			ID:         getSnapshotResponse.Snapshot.ID,
			Name:       getSnapshotResponse.Snapshot.Name,
			VolumeType: getSnapshotResponse.Snapshot.VolumeType,
		}

		return snap, nil
	}

	blockSnapshot, err := api.BlockAPI.GetSnapshot(&block.GetSnapshotRequest{
		Zone:       req.Zone,
		SnapshotID: req.SnapshotID,
	}, opts...)
	if err != nil {
		return nil, err
	}

	snap := &UnknownSnapshot{
		Zone:       blockSnapshot.Zone,
		ID:         blockSnapshot.ID,
		Name:       blockSnapshot.Name,
		VolumeType: instance.VolumeVolumeTypeSbsSnapshot,
	}

	return snap, nil
}
