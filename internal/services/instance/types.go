package instance

import (
	"strconv"

	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
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

func flattenImageExtraVolumes(volumes map[string]*instance.Volume, zone scw.Zone) interface{} {
	volumesFlat := []map[string]interface{}(nil)

	for _, volume := range volumes {
		server := map[string]interface{}{}
		if volume.Server != nil {
			server["id"] = volume.Server.ID
			server["name"] = volume.Server.Name
		}

		volumeFlat := map[string]interface{}{
			"id":                zonal.NewIDString(zone, volume.ID),
			"name":              volume.Name,
			"export_uri":        volume.ExportURI,
			"size":              volume.Size,
			"volume_type":       volume.VolumeType,
			"creation_date":     volume.CreationDate,
			"modification_date": volume.ModificationDate,
			"organization":      volume.Organization,
			"project":           volume.Project,
			"tags":              volume.Tags,
			"state":             volume.State,
			"zone":              volume.Zone,
			"server":            server,
		}
		volumesFlat = append(volumesFlat, volumeFlat)
	}

	return volumesFlat
}

func flattenServerPublicIPs(zone scw.Zone, ips []*instance.ServerIP) []interface{} {
	flattenedIPs := make([]interface{}, len(ips))

	for i, ip := range ips {
		flattenedIPs[i] = map[string]interface{}{
			"id":      zonal.NewIDString(zone, ip.ID),
			"address": ip.Address.String(),
		}
	}

	return flattenedIPs
}

func flattenServerIPIDs(ips []*instance.ServerIP) []interface{} {
	ipIDs := make([]interface{}, len(ips))

	for i, ip := range ips {
		ipIDs[i] = ip.ID
	}

	return ipIDs
}
