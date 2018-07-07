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
	mu.Lock()
	task, err := scaleway.PostServerAction(server.Identifier, "terminate")
	mu.Unlock()

	if err != nil {
		if serr, ok := err.(api.APIError); ok {
			if serr.StatusCode == 404 {
				return nil
			}
		}

		return err
	}

	return waitForTaskCompletion(scaleway, task.Identifier)
}

// deleteStoppedServer needs to cleanup attached root volumes. this is not done
// automatically by Scaleway
func deleteStoppedServer(scaleway *api.API, server *api.Server) error {
	mu.Lock()
	defer mu.Unlock()
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

func waitForTaskCompletion(scaleway *api.API, taskID string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"pending", "started"},
		Target:  []string{"success"},
		Refresh: func() (interface{}, string, error) {
			mu.Lock()
			defer mu.Unlock()

			task, err := scaleway.GetTask(taskID)
			if err != nil {
				return 42, "error", err
			}
			return 42, task.Status, nil
		},
		Timeout:    60 * time.Minute,
		MinTimeout: 10 * time.Second,
		Delay:      15 * time.Second,
	}
	_, err := stateConf.WaitForState()
	return err
}

func withStoppedServer(scaleway *api.API, serverID string, run func(*api.Server) error) error {
	server, err := scaleway.GetServer(serverID)
	if err != nil {
		return err
	}

	var startServerAgain = false

	if server.State != "stopped" {
		startServerAgain = true

		task, err := scaleway.PostServerAction(server.Identifier, "poweroff")
		if err != nil {
			return err
		}
		if err := waitForTaskCompletion(scaleway, task.Identifier); err != nil {
			return err
		}
	}

	if err := run(server); err != nil {
		return err
	}

	if startServerAgain {
		task, err := scaleway.PostServerAction(serverID, "poweron")
		if err != nil {
			return err
		}
		if err := waitForTaskCompletion(scaleway, task.Identifier); err != nil {
			return err
		}
	}
	return nil
}
