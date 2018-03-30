package scaleway

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	api "github.com/nicolai86/scaleway-sdk"
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
func deleteRunningServer(scaleway *api.API, server *api.Server) error {
	_, err := scaleway.PostServerAction(server.Identifier, "terminate")

	if err != nil {
		if serr, ok := err.(api.APIError); ok {
			if serr.StatusCode == 404 {
				return nil
			}
		}

		return err
	}

	return waitForServerShutdown(scaleway, server.Identifier)
}

// deleteStoppedServer needs to cleanup attached root volumes. this is not done
// automatically by Scaleway
func deleteStoppedServer(scaleway *api.API, server *api.Server) error {
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

func waitForServerStartup(scaleway *api.API, serverID string) error {
	return waitForServerState(scaleway, serverID, "running", []string{"running", "starting"})
}

func waitForServerShutdown(scaleway *api.API, serverID string) error {
	return waitForServerState(scaleway, serverID, "stopped", []string{"stopped", "stopping"})
}

func waitForServerState(scaleway *api.API, serverID, targetState string, pendingStates []string) error {
	stateConf := &resource.StateChangeConf{
		Pending: pendingStates,
		Target:  []string{targetState},
		Refresh: func() (interface{}, string, error) {
			s, err := scaleway.GetServer(serverID)

			if err == nil {
				return 42, s.State, nil
			}

			if serr, ok := err.(api.APIError); ok {
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
		MinTimeout: 10 * time.Second,
		Delay:      15 * time.Second,
	}
	_, err := stateConf.WaitForState()
	return err
}
