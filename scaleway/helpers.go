package scaleway

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/nicolai86/scaleway-sdk/api"
)

// Bool returns a pointer to of the bool value passed in.
func Bool(val bool) *bool {
	return &val
}

// String returns a pointer to of the string value passed in.
func String(val string) *string {
	return &val
}

func validateServerType(v interface{}, k string) (ws []string, errors []error) {
	// only validate if we were able to fetch a list of commercial types
	if len(commercialServerTypes) == 0 {
		return
	}

	isKnown := false
	requestedType := v.(string)
	for _, knownType := range commercialServerTypes {
		isKnown = isKnown || strings.ToUpper(knownType) == strings.ToUpper(requestedType)
	}

	if !isKnown {
		errors = append(errors, fmt.Errorf("%q must be one of %q", k, commercialServerTypes))
	}
	return
}

func validateVolumeType(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if value != "l_ssd" {
		errors = append(errors, fmt.Errorf("%q must be l_ssd", k))
	}
	return
}

// deleteRunningServer terminates the server and waits until it is removed.
func deleteRunningServer(scaleway *api.ScalewayAPI, server *api.ScalewayServer) error {
	err := scaleway.PostServerAction(server.Identifier, "terminate")

	if err != nil {
		if serr, ok := err.(api.ScalewayAPIError); ok {
			if serr.StatusCode == 404 {
				return nil
			}
		}

		return err
	}

	return waitForServerState(scaleway, server.Identifier, "stopped")
}

// deleteStoppedServer needs to cleanup attached root volumes. this is not done
// automatically by Scaleway
func deleteStoppedServer(scaleway *api.ScalewayAPI, server *api.ScalewayServer) error {
	if err := scaleway.DeleteServer(server.Identifier); err != nil {
		return err
	}

	if rootVolume, ok := server.Volumes["0"]; ok {
		if err := scaleway.DeleteVolume(rootVolume.Identifier); err != nil {
			return err
		}
	}
	return nil
}

// NOTE copied from github.com/scaleway/scaleway-cli/pkg/api/helpers.go
// the helpers.go file pulls in quite a lot dependencies, and they're just convenience wrappers anyway

var allStates = []string{"starting", "running", "stopping", "stopped"}

func waitForServerState(scaleway *api.ScalewayAPI, serverID, targetState string) error {
	pending := []string{}
	for _, state := range allStates {
		if state != targetState {
			pending = append(pending, state)
		}
	}
	stateConf := &resource.StateChangeConf{
		Pending: pending,
		Target:  []string{targetState},
		Refresh: func() (interface{}, string, error) {
			s, err := scaleway.GetServer(serverID)

			if err == nil {
				return 42, s.State, nil
			}

			if serr, ok := err.(api.ScalewayAPIError); ok {
				if serr.StatusCode == 404 {
					return 42, "stopped", nil
				}
			}

			if s != nil {
				return 42, s.State, err
			}
			return 42, "error", err
		},
		Timeout:    60 * time.Minute,
		MinTimeout: 5 * time.Second,
		Delay:      5 * time.Second,
	}
	_, err := stateConf.WaitForState()
	return err
}
