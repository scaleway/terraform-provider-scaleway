package scaleway

import (
	"context"
	"fmt"
	"hash/crc32"
	"sort"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	InstanceServerStateStopped = "stopped"
	InstanceServerStateStarted = "started"
	InstanceServerStateStandby = "standby"

	InstanceServerWaitForTimeout = 10 * time.Minute
)

// instanceAPIWithZone returns a new instance API and the zone for a Create request
func instanceAPIWithZone(d *schema.ResourceData, m interface{}) (*instance.API, scw.Zone, error) {
	meta := m.(*Meta)
	instanceAPI := instance.NewAPI(meta.scwClient)

	zone, err := extractZone(d, meta)
	return instanceAPI, zone, err
}

// instanceAPIWithZoneAndID returns an instance API with zone and ID extracted from the state
func instanceAPIWithZoneAndID(m interface{}, zonedID string) (*instance.API, scw.Zone, string, error) {
	meta := m.(*Meta)
	instanceAPI := instance.NewAPI(meta.scwClient)

	zone, ID, err := parseZonedID(zonedID)
	return instanceAPI, zone, ID, err
}

// instanceAPIWithZoneAndNestedID returns an instance API with zone and inner/outer ID extracted from the state
func instanceAPIWithZoneAndNestedID(m interface{}, zonedNestedID string) (*instance.API, scw.Zone, string, string, error) {
	meta := m.(*Meta)
	instanceAPI := instance.NewAPI(meta.scwClient)

	zone, innerID, outerID, err := parseZonedNestedID(zonedNestedID)
	return instanceAPI, zone, innerID, outerID, err
}

// hash hashes a string to a unique hashcode.
//
// crc32 returns a uint32, but for our use we need
// and non negative integer. Here we cast to an integer
// and invert it if the result is negative.
func hash(s string) int {
	v := int(crc32.ChecksumIEEE([]byte(s)))
	if v >= 0 {
		return v
	}
	if -v >= 0 {
		return -v
	}
	// v == MinInt
	return 0
}

func userDataHash(v interface{}) int {
	userData := v.(map[string]interface{})
	return hash(userData["key"].(string) + userData["value"].(string))
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

	for _, a := range actions {
		err = instanceAPI.ServerActionAndWait(&instance.ServerActionAndWaitRequest{
			ServerID: serverID,
			Action:   a,
			Zone:     zone,
			Timeout:  scw.TimeDurationPtr(InstanceServerWaitForTimeout),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// detachVolume will make sure a volume is not attached to any server. If volume is attached to a server, it will be stopped
// to allow volume detachment.
func detachVolume(ctx context.Context, instanceAPI *instance.API, zone scw.Zone, volumeID string) error {
	res, err := instanceAPI.GetVolume(&instance.GetVolumeRequest{
		Zone:     zone,
		VolumeID: volumeID,
	}, scw.WithContext(ctx))
	if err != nil {
		return err
	}

	if res.Volume.Server == nil {
		return nil
	}

	defer lockLocalizedID(newZonedIDString(zone, res.Volume.Server.ID))()

	// We need to stop server only for VolumeTypeLSSD volume type
	if res.Volume.VolumeType == instance.VolumeVolumeTypeLSSD {
		err = reachState(ctx, instanceAPI, zone, res.Volume.Server.ID, instance.ServerStateStopped)

		// If 404 this mean server is deleted and volume is already detached
		if is404Error(err) {
			return nil
		}
		if err != nil {
			return err
		}
	}
	_, err = instanceAPI.DetachVolume(&instance.DetachVolumeRequest{
		Zone:     zone,
		VolumeID: res.Volume.ID,
	}, scw.WithContext(ctx))

	// TODO find a better way to test this error
	if err != nil && err.Error() != "scaleway-sdk-go: volume should be attached to a server" {
		return err
	}

	return nil
}
