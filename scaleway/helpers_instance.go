package scaleway

import (
	"context"
	"fmt"
	"github.com/scaleway/scaleway-sdk-go/api/vpc/v1"
	"sort"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	InstanceServerStateStopped = "stopped"
	InstanceServerStateStarted = "started"
	InstanceServerStateStandby = "standby"

	defaultInstanceServerWaitTimeout        = 10 * time.Minute
	defaultInstanceVolumeDeleteTimeout      = 10 * time.Minute
	defaultInstanceSecurityGroupTimeout     = 1 * time.Minute
	defaultInstanceSecurityGroupRuleTimeout = 1 * time.Minute
	defaultInstancePlacementGroupTimeout    = 1 * time.Minute
	defaultInstanceIPTimeout                = 1 * time.Minute

	defaultInstanceSnapshotWaitTimeout = 1 * time.Hour
)

// instanceAPIWithZone returns a new instance API and the zone for a Create request
func instanceAPIWithZone(d *schema.ResourceData, m interface{}) (*instance.API, scw.Zone, error) {
	meta := m.(*Meta)
	instanceAPI := instance.NewAPI(meta.scwClient)

	zone, err := extractZone(d, meta)
	if err != nil {
		return nil, "", err
	}
	return instanceAPI, zone, nil
}

// instanceAPIWithZoneAndID returns an instance API with zone and ID extracted from the state
func instanceAPIWithZoneAndID(m interface{}, zonedID string) (*instance.API, scw.Zone, string, error) {
	meta := m.(*Meta)
	instanceAPI := instance.NewAPI(meta.scwClient)

	zone, ID, err := parseZonedID(zonedID)
	if err != nil {
		return nil, "", "", err
	}
	return instanceAPI, zone, ID, nil
}

// instanceAPIWithZoneAndNestedID returns an instance API with zone and inner/outer ID extracted from the state
func instanceAPIWithZoneAndNestedID(m interface{}, zonedNestedID string) (*instance.API, scw.Zone, string, string, error) {
	meta := m.(*Meta)
	instanceAPI := instance.NewAPI(meta.scwClient)

	zone, innerID, outerID, err := parseZonedNestedID(zonedNestedID)
	if err != nil {
		return nil, "", "", "", err
	}
	return instanceAPI, zone, innerID, outerID, nil
}

// orderVolumes return an ordered slice based on the volume map key "0", "1", "2",...
func orderVolumes(v map[string]*instance.Volume) []*instance.Volume {
	indexes := []string{}
	for index := range v {
		indexes = append(indexes, index)
	}
	sort.Strings(indexes)
	var orderedVolumes []*instance.Volume
	for _, index := range indexes {
		orderedVolumes = append(orderedVolumes, v[index])
	}
	return orderedVolumes
}

// serverStateFlatten converts the API state to terraform state or return an error.
func serverStateFlatten(fromState instance.ServerState) (string, error) {
	switch fromState {
	case instance.ServerStateStopped:
		return InstanceServerStateStopped, nil
	case instance.ServerStateStoppedInPlace:
		return InstanceServerStateStandby, nil
	case instance.ServerStateRunning:
		return InstanceServerStateStarted, nil
	case instance.ServerStateLocked:
		return "", fmt.Errorf("server is locked, please contact Scaleway support: https://console.scaleway.com/support/tickets")
	}
	return "", fmt.Errorf("server is in an invalid state, someone else might be executing action at the same time")
}

// serverStateExpand converts a terraform state  to an API state or return an error.
func serverStateExpand(rawState string) (instance.ServerState, error) {
	apiState, exist := map[string]instance.ServerState{
		InstanceServerStateStopped: instance.ServerStateStopped,
		InstanceServerStateStandby: instance.ServerStateStoppedInPlace,
		InstanceServerStateStarted: instance.ServerStateRunning,
	}[rawState]

	if !exist {
		return "", fmt.Errorf("server is in a transient state, someone else might be executing another action at the same time")
	}

	return apiState, nil
}

func reachState(ctx context.Context, instanceAPI *instance.API, zone scw.Zone, serverID string, toState instance.ServerState) error {
	response, err := instanceAPI.GetServer(&instance.GetServerRequest{
		Zone:     zone,
		ServerID: serverID,
	}, scw.WithContext(ctx))
	if err != nil {
		return err
	}
	fromState := response.Server.State

	if response.Server.State == toState {
		return nil
	}

	transitionMap := map[[2]instance.ServerState][]instance.ServerAction{
		{instance.ServerStateStopped, instance.ServerStateRunning}:        {instance.ServerActionPoweron},
		{instance.ServerStateStopped, instance.ServerStateStoppedInPlace}: {instance.ServerActionPoweron, instance.ServerActionStopInPlace},
		{instance.ServerStateRunning, instance.ServerStateStopped}:        {instance.ServerActionPoweroff},
		{instance.ServerStateRunning, instance.ServerStateStoppedInPlace}: {instance.ServerActionStopInPlace},
		{instance.ServerStateStoppedInPlace, instance.ServerStateRunning}: {instance.ServerActionPoweron},
		{instance.ServerStateStoppedInPlace, instance.ServerStateStopped}: {instance.ServerActionPoweron, instance.ServerActionPoweroff},
	}

	actions, exist := transitionMap[[2]instance.ServerState{fromState, toState}]
	if !exist {
		return fmt.Errorf("don't know how to reach state %s from state %s for server %s", toState, fromState, serverID)
	}

	// We need to check that all volumes are ready
	for _, volume := range response.Server.Volumes {
		if volume.State != instance.VolumeStateAvailable {
			_, err = instanceAPI.WaitForVolume(&instance.WaitForVolumeRequest{
				Zone:          zone,
				VolumeID:      volume.ID,
				RetryInterval: DefaultWaitRetryInterval,
			})
			if err != nil {
				return err
			}
		}
	}

	for _, a := range actions {
		err = instanceAPI.ServerActionAndWait(&instance.ServerActionAndWaitRequest{
			ServerID:      serverID,
			Action:        a,
			Zone:          zone,
			Timeout:       scw.TimeDurationPtr(defaultInstanceServerWaitTimeout),
			RetryInterval: DefaultWaitRetryInterval,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// getServerType is a util to get a instance.ServerType by its commercialType
func getServerType(apiInstance *instance.API, zone scw.Zone, commercialType string) *instance.ServerType {
	serverType := (*instance.ServerType)(nil)

	serverTypesRes, err := apiInstance.ListServersTypes(&instance.ListServersTypesRequest{
		Zone: zone,
	})
	if err != nil {
		l.Warningf("cannot get server types: %s", err)
	} else {
		serverType = serverTypesRes.Servers[commercialType]
		if serverType == nil {
			l.Warningf("unrecognized server type: %s", commercialType)
		}
	}

	return serverType
}

// validateLocalVolumeSizes validates the total size of local volumes.
func validateLocalVolumeSizes(volumes map[string]*instance.VolumeTemplate, serverType *instance.ServerType, commercialType string) error {
	// Calculate local volume total size.
	var localVolumeTotalSize scw.Size
	for _, volume := range volumes {
		if volume.VolumeType == instance.VolumeVolumeTypeLSSD {
			localVolumeTotalSize += volume.Size
		}
	}

	volumeConstraint := serverType.VolumesConstraint

	// If no root volume provided, count the default root volume size added by the API.
	if rootVolume := volumes["0"]; rootVolume == nil {
		localVolumeTotalSize += volumeConstraint.MinSize
	}

	if localVolumeTotalSize < volumeConstraint.MinSize || localVolumeTotalSize > volumeConstraint.MaxSize {
		min := humanize.Bytes(uint64(volumeConstraint.MinSize))
		if volumeConstraint.MinSize == volumeConstraint.MaxSize {
			return fmt.Errorf("%s total local volume size must be equal to %s", commercialType, min)
		}

		max := humanize.Bytes(uint64(volumeConstraint.MaxSize))
		return fmt.Errorf("%s total local volume size must be between %s and %s", commercialType, min, max)
	}

	return nil
}

// sanitizeVolumeMap removes extra data for API validation.
//
// On the api side, there are two possibles validation schemas for volumes and the validator will be chosen dynamically depending on the passed JSON request
// - With an image (in that case the root volume can be skipped because it is taken from the image)
// - Without an image (in that case, the root volume must be defined)
func sanitizeVolumeMap(serverName string, volumes map[string]*instance.VolumeTemplate) map[string]*instance.VolumeTemplate {
	m := make(map[string]*instance.VolumeTemplate)

	for index, v := range volumes {
		v.Name = serverName + "-" + index

		// Remove extra data for API validation.
		switch {
		// If a volume already got an ID it is passed as it to the API without specifying the volume type.
		// TODO: Fix once instance accept volume type in the schema validation
		case v.ID != "":
			v = &instance.VolumeTemplate{ID: v.ID, Name: v.Name}
		// For the root volume (index 0) if the specified size is not 0 it is considered as a new volume
		// It does not have yet a volume ID, it is passed to the API with only the size to be dynamically created by the API
		case index == "0" && v.Size != 0:
			v = &instance.VolumeTemplate{Size: v.Size}
		// If none of the above conditions are met, the volume is passed as it to the API
		default:
		}
		m[index] = v
	}

	return m
}

func preparePrivateNIC(
	ctx context.Context, data interface{},
	server *instance.Server, vpcAPI *vpc.API) ([]*instance.CreatePrivateNICRequest, error) {
	if data == nil {
		return nil, nil
	}

	var res []*instance.CreatePrivateNICRequest

	for _, pn := range data.([]interface{}) {
		r := pn.(map[string]interface{})
		zonedID, pnExist := r["pn_id"]
		privateNetworkID := expandID(zonedID.(string))
		if pnExist {
			currentPN, err := vpcAPI.GetPrivateNetwork(&vpc.GetPrivateNetworkRequest{
				PrivateNetworkID: expandID(privateNetworkID),
				Zone:             server.Zone,
			}, scw.WithContext(ctx))
			if err != nil {
				return nil, err
			}
			query := &instance.CreatePrivateNICRequest{
				Zone: currentPN.Zone, ServerID: server.ID, PrivateNetworkID: currentPN.ID}
			res = append(res, query)
		}
	}

	return res, nil
}

func privateNICFlatten(l *instance.ListPrivateNICsResponse, zone scw.Zone) (interface{}, error) {
	if l == nil {
		return nil, nil
	}
	privateNetworks := []map[string]interface{}(nil)
	for _, pn := range l.PrivateNics {
		privateNetworks = append(privateNetworks, map[string]interface{}{
			"pn_id":       newZonedID(zone, pn.PrivateNetworkID).String(),
			"mac_address": pn.MacAddress,
			"status":      pn.State.String(),
			"zone":        zone.String(),
		})
	}

	return privateNetworks, nil
}
