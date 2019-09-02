package scaleway

import (
	"fmt"
	"sort"
	"time"

	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	InstanceServerStateStopped = "stopped"
	InstanceServerStateStarted = "started"
	InstanceServerStateStandby = "standby"

	InstanceServerWaitForTimeout   = 10 * time.Minute
	InstanceServerRetryFuncTimeout = InstanceServerWaitForTimeout + time.Minute // some RetryFunc are calling a WaitFor
)

// getInstanceAPIWithZone returns a new instance API and the zone for a Create request
func getInstanceAPIWithZone(d *schema.ResourceData, m interface{}) (*instance.API, scw.Zone, error) {
	meta := m.(*Meta)
	instanceAPI := instance.NewAPI(meta.scwClient)

	zone, err := getZone(d, meta)
	return instanceAPI, zone, err
}

// getInstanceAPIWithZoneAndID returns an instance API with zone and ID extracted from the state
func getInstanceAPIWithZoneAndID(m interface{}, id string) (*instance.API, scw.Zone, string, error) {
	meta := m.(*Meta)
	instanceAPI := instance.NewAPI(meta.scwClient)

	zone, ID, err := parseZonedID(id)
	return instanceAPI, zone, ID, err
}

func userDataHash(v interface{}) int {
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

// instanceServerStateToAction returns the action required to transit from a state to another.
func instanceServerStateToAction(previousState, nextState string, forceReboot bool) []instance.ServerAction {
	if previousState == InstanceServerStateStarted && nextState == InstanceServerStateStarted && forceReboot {
		return []instance.ServerAction{instance.ServerActionReboot}
	}
	transitionMap := map[[2]string][]instance.ServerAction{
		{InstanceServerStateStopped, InstanceServerStateStopped}: {},
		{InstanceServerStateStopped, InstanceServerStateStarted}: {instance.ServerActionPoweron},
		{InstanceServerStateStopped, InstanceServerStateStandby}: {instance.ServerActionPoweron, instance.ServerActionStopInPlace},
		{InstanceServerStateStarted, InstanceServerStateStopped}: {instance.ServerActionPoweroff},
		{InstanceServerStateStarted, InstanceServerStateStarted}: {},
		{InstanceServerStateStarted, InstanceServerStateStandby}: {instance.ServerActionStopInPlace},
		{InstanceServerStateStandby, InstanceServerStateStopped}: {instance.ServerActionPoweroff},
		{InstanceServerStateStandby, InstanceServerStateStarted}: {instance.ServerActionPoweron},
		{InstanceServerStateStandby, InstanceServerStateStandby}: {},
	}

	return transitionMap[[2]string{previousState, nextState}]
}

// reachState executes server action(s) to reach the expected state
func reachState(instanceAPI *instance.API, zone scw.Zone, serverID, fromState, toState string, forceReboot bool) error {
	for _, action := range instanceServerStateToAction(fromState, toState, forceReboot) {
		err := resource.Retry(InstanceServerRetryFuncTimeout, func() *resource.RetryError {
			err := instanceAPI.ServerActionAndWait(&instance.ServerActionAndWaitRequest{
				Zone:     zone,
				ServerID: serverID,
				Action:   action,
				Timeout:  InstanceServerWaitForTimeout,
			})
			if isSDKError(err, "expected state [\\w]+ but found [\\w]+") {
				l.Errorf("Retrying action %s because of error '%v'", action, err)
				return resource.RetryableError(err)
			}
			if err != nil {
				return resource.NonRetryableError(err)
			}
			return nil
		})
		if err != nil {
			return err
		}

	}
	return nil
}
