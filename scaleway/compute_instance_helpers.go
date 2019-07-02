package scaleway

import (
	"fmt"
	"sort"
	"time"

	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/utils"
)

const (
	ServerStateStopped = "stopped"
	ServerStateStarted = "started"
	ServerStateStandby = "standby"

	ServerWaitForTimeout = 10 * time.Minute
)

// getInstanceAPIWithZone returns a new instance API and the zone for a Create request
func getInstanceAPIWithZone(d *schema.ResourceData, m interface{}) (*instance.API, utils.Zone, error) {
	meta := m.(*Meta)
	instanceApi := instance.NewAPI(meta.scwClient)

	zone, err := getZone(d, meta)
	return instanceApi, zone, err
}

// getInstanceAPIWithZoneAndID returns an instance API with zone and ID extracted from the state
func getInstanceAPIWithZoneAndID(m interface{}, id string) (*instance.API, utils.Zone, string, error) {
	meta := m.(*Meta)
	instanceApi := instance.NewAPI(meta.scwClient)

	zone, ID, err := parseZonedID(id)
	return instanceApi, zone, ID, err
}

// expandRootVolume expands the current root volume or returns the default root volume
func expandRootVolume(v interface{}) map[string]interface{} {
	flattenVolume := map[string]interface{}{}

	vs, ok := v.([]map[string]interface{})
	if ok && len(vs) > 0 {
		flattenVolume = vs[0]
	}

	if _, exist := flattenVolume["delete_on_termination"]; !exist {
		flattenVolume["delete_on_termination"] = true // default value does not work on list
	}

	return flattenVolume
}

func schemaSetUserData(v interface{}) int {
	userData := v.(map[string]interface{})
	return hashcode.String(userData["key"].(string) + userData["value"].(string))
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

// expandServerState converts the API state to terraform state or return an error.
func expandServerState(fromState instance.ServerState) (string, error) {
	switch fromState {
	case instance.ServerStateStopped:
		return ServerStateStopped, nil
	case instance.ServerStateStoppedInPlace:
		return ServerStateStandby, nil
	case instance.ServerStateRunning:
		return ServerStateStarted, nil
	case instance.ServerStateLocked:
		return "", fmt.Errorf("server is locked, please contact Scaleway support: https://console.scaleway.com/support/tickets")
	}
	return "", fmt.Errorf("server is in an invalid state, someone else might be executing action at the same time")
}

// computeServerStateToAction returns the action required to transit from a state to another.
func computeServerStateToAction(previousState, nextState string, forceReboot bool) []instance.ServerAction {
	if previousState == ServerStateStarted && nextState == ServerStateStarted && forceReboot {
		return []instance.ServerAction{instance.ServerActionReboot}
	}
	transitionMap := map[[2]string][]instance.ServerAction{
		{ServerStateStopped, ServerStateStopped}: {},
		{ServerStateStopped, ServerStateStarted}: {instance.ServerActionPoweron},
		{ServerStateStopped, ServerStateStandby}: {instance.ServerActionPoweron, instance.ServerActionStopInPlace},
		{ServerStateStarted, ServerStateStopped}: {instance.ServerActionPoweroff},
		{ServerStateStarted, ServerStateStarted}: {},
		{ServerStateStarted, ServerStateStandby}: {instance.ServerActionStopInPlace},
		{ServerStateStandby, ServerStateStopped}: {instance.ServerActionPoweroff},
		{ServerStateStandby, ServerStateStarted}: {instance.ServerActionPoweron},
		{ServerStateStandby, ServerStateStandby}: {},
	}

	return transitionMap[[2]string{previousState, nextState}]
}

// reachState executes server action(s) to reach the expected state
func reachState(instanceApi *instance.API, zone utils.Zone, serverID, fromState, toState string, forceReboot bool) error {
	for _, action := range computeServerStateToAction(fromState, toState, forceReboot) {
		err := instanceApi.ServerActionAndWait(&instance.ServerActionAndWaitRequest{
			Zone:     zone,
			ServerID: serverID,
			Action:   action,
			Timeout:  ServerWaitForTimeout,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
