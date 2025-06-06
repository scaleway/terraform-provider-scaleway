package instance

import (
	"strconv"

	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
)

func (ph *privateNICsHandler) flatPrivateNICs() error {
	privateNICsMap := make(map[string]*instance.PrivateNIC)

	res, err := ph.instanceAPI.ListPrivateNICs(&instance.ListPrivateNICsRequest{Zone: ph.zone, ServerID: ph.serverID})
	if err != nil {
		return err
	}

	for _, p := range res.PrivateNics {
		privateNICsMap[p.PrivateNetworkID] = p
	}

	ph.privateNICsMap = privateNICsMap

	return nil
}

func expandImageExtraVolumesTemplates(snapshotIDs []string) map[string]*instance.VolumeTemplate {
	volTemplates := map[string]*instance.VolumeTemplate{}
	if len(snapshotIDs) == 0 {
		return volTemplates
	}

	for i, snapID := range snapshotIDs {
		volTemplate := &instance.VolumeTemplate{
			ID: snapID,
		}
		volTemplates[strconv.Itoa(i+1)] = volTemplate
	}

	return volTemplates
}

func expandImageExtraVolumesUpdateTemplates(snapshotIDs []string) map[string]*instance.VolumeImageUpdateTemplate {
	volTemplates := map[string]*instance.VolumeImageUpdateTemplate{}
	if len(snapshotIDs) == 0 {
		return volTemplates
	}

	for i, snapID := range snapshotIDs {
		volTemplate := &instance.VolumeImageUpdateTemplate{
			ID: snapID,
		}
		volTemplates[strconv.Itoa(i+1)] = volTemplate
	}

	return volTemplates
}

func flattenImageRootVolume(volume *instance.VolumeSummary, zone scw.Zone) any {
	volumeFlat := map[string]any{
		"id":          zonal.NewIDString(zone, volume.ID),
		"name":        volume.Name,
		"size":        volume.Size,
		"volume_type": volume.VolumeType,
	}

	return []map[string]any{volumeFlat}
}

func flattenImageExtraVolumes(volumes map[string]*instance.Volume, zone scw.Zone) any {
	volumesFlat := []map[string]any(nil)

	for index := 1; index < len(volumes)+1; index++ {
		volume := volumes[strconv.Itoa(index)]

		volumeFlat := map[string]any{
			"id":          zonal.NewIDString(zone, volume.ID),
			"name":        volume.Name,
			"size":        volume.Size,
			"volume_type": volume.VolumeType,
			"tags":        volume.Tags,
		}
		if volume.Server != nil {
			server := map[string]any{}
			server["id"] = volume.Server.ID
			server["name"] = volume.Server.Name
			volumeFlat["server"] = server
		}

		volumesFlat = append(volumesFlat, volumeFlat)
	}

	return volumesFlat
}

func flattenServerPublicIPs(zone scw.Zone, ips []*instance.ServerIP) []any {
	flattenedIPs := make([]any, len(ips))

	for i, ip := range ips {
		flattenedIPs[i] = map[string]any{
			"id":                zonal.NewIDString(zone, ip.ID),
			"address":           ip.Address.String(),
			"gateway":           ip.Gateway.String(),
			"netmask":           ip.Netmask,
			"family":            ip.Family.String(),
			"dynamic":           ip.Dynamic,
			"provisioning_mode": ip.ProvisioningMode.String(),
		}
	}

	return flattenedIPs
}

func flattenServerFileSystem(zone scw.Zone, fs []*instance.ServerFilesystem) []any {
	filesystems := make([]any, len(fs))
	region, _ := zone.Region()

	for i, f := range fs {
		filesystems[i] = map[string]any{
			"filesystem_id": regional.NewIDString(region, f.FilesystemID),
			"status":        f.State,
		}
	}

	return filesystems
}

func flattenServerIPIDs(ips []*instance.ServerIP) []any {
	ipIDs := make([]any, len(ips))

	for i, ip := range ips {
		ipIDs[i] = ip.ID
	}

	return ipIDs
}
