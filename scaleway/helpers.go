package scaleway

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	api "github.com/nicolai86/scaleway-sdk"
	"github.com/scaleway/scaleway-sdk-go/namegenerator"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/scaleway-sdk-go/utils"
	"golang.org/x/xerrors"
)

// userAgent used for SDK requests.
var userAgent = fmt.Sprintf("terraform-provider/%s terraform/%s", version, terraform.VersionString())

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

var allStates = []string{"starting", "running", "stopping", "stopped"}

func waitForServerShutdown(scaleway *api.API, serverID string) error {
	return waitForServerState(scaleway, serverID, "stopped", []string{"stopped", "stopping"})
}

func waitForServerStartup(scaleway *api.API, serverID string) error {
	return waitForServerState(scaleway, serverID, "running", []string{"running", "starting"})
}

func waitForServerState(scaleway *api.API, serverID, targetState string, pendingStates []string) error {
	wg := getWaitForServerLock(serverID)
	wg.Wait()

	mu.Lock()
	wg.Add(1)
	mu.Unlock()

	defer func() {
		mu.Lock()
		wg.Done()
		mu.Unlock()
	}()

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

var waitForServer = map[string]*sync.WaitGroup{}

func getWaitForServerLock(serverID string) *sync.WaitGroup {
	mu.Lock()
	defer mu.Unlock()
	wg, ok := waitForServer[serverID]
	if !ok {
		wg = &sync.WaitGroup{}
		waitForServer[serverID] = wg
	}
	return wg
}

func startServer(scaleway *api.API, server *api.Server) error {
	wg := getWaitForServerLock(server.Identifier)
	wg.Wait()

	_, err := scaleway.PostServerAction(server.Identifier, "poweron")

	if err != nil {
		return err
	}

	return waitForServerStartup(scaleway, server.Identifier)
}

func stopServer(scaleway *api.API, server *api.Server) error {
	wg := getWaitForServerLock(server.Identifier)
	wg.Wait()

	_, err := scaleway.PostServerAction(server.Identifier, "poweroff")

	if err != nil {
		return err
	}
	return waitForServerShutdown(scaleway, server.Identifier)
}

// deleteRunningServer terminates the server and waits until it is removed.
func deleteRunningServer(scaleway *api.API, server *api.Server) error {
	wg := getWaitForServerLock(server.Identifier)
	wg.Wait()

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

func withStoppedServer(scaleway *api.API, serverID string, run func(*api.Server) error) error {
	wg := getWaitForServerLock(serverID)
	wg.Wait()

	server, err := scaleway.GetServer(serverID)

	if err != nil {
		return err
	}

	var startServerAgain = false
	if server.State != "stopped" {
		startServerAgain = true

		err := stopServer(scaleway, server)
		if err != nil {
			return err
		}
	}

	if err := run(server); err != nil {
		return err
	}

	if startServerAgain {
		err := startServer(scaleway, server)
		if err != nil {
			return err
		}
	}
	return nil
}

// parseLocalizedID parses a localizedID and extracts the resource locality and id.
func parseLocalizedID(localizedID string) (locality string, ID string, err error) {
	tab := strings.SplitN(localizedID, "/", 2)
	if len(tab) != 2 {
		return "", "", fmt.Errorf("cant parse localized id: %s", localizedID)
	}
	return tab[0], tab[1], nil
}

// parseZonedID parses a zonedID and extracts the resource zone and id.
func parseZonedID(zonedID string) (zone utils.Zone, id string, err error) {

	locality, id, err := parseLocalizedID(zonedID)
	if err != nil {
		return
	}

	zone, err = utils.ParseZone(locality)
	return
}

// parseRegionalID parses a regionalID and extracts the resource region and id.
func parseRegionalID(regionalID string) (region utils.Region, id string, err error) {
	locality, id, err := parseLocalizedID(regionalID)
	if err != nil {
		return
	}

	region, err = utils.ParseRegion(locality)
	return
}

// newZonedId constructs a unique identifier based on resource zone and id
func newZonedId(zone utils.Zone, id string) string {
	return fmt.Sprintf("%s/%s", zone, id)
}

// newRegionalId constructs a unique identifier based on resource region and id
func newRegionalId(region utils.Region, id string) string {
	return fmt.Sprintf("%s/%s", region, id)
}

// terraformResourceData is an interface for *schema.ResourceData. (used for mock)
type terraformResourceData interface {
	HasChange(string) bool
	GetOkExists(string) (interface{}, bool)
	GetOk(string) (interface{}, bool)
	Get(string) interface{}
	Set(string, interface{}) error
	SetId(string)
	Id() string
}

// getZone will try to guess the zone from the following:
//  - zone field of the resource data
//  - default zone from config
func getZone(d terraformResourceData, meta *Meta) (utils.Zone, error) {

	rawZone, exist := d.GetOkExists("zone")
	if exist {
		return utils.ParseZone(rawZone.(string))
	}

	zone, exist := meta.scwClient.GetDefaultZone()
	if exist {
		return zone, nil
	}

	return utils.Zone(""), fmt.Errorf("could not detect region")
}

// isHTTPCodeError returns true if err is an http error with code statusCode
func isHTTPCodeError(err error, statusCode int) bool {
	if err == nil {
		return false
	}

	responseError := &scw.ResponseError{}
	if xerrors.As(err, &responseError) && responseError.StatusCode == statusCode {
		return true
	}
	return false
}

// is404Error returns true if err is an HTTP 404 error
func is404Error(err error) bool {
	return isHTTPCodeError(err, http.StatusNotFound)
}

// is403Error returns true if err is an HTTP 403 error
func is403Error(err error) bool {
	return isHTTPCodeError(err, http.StatusForbidden)
}

// projectIDSchema returns a standard schema for a project_id
func projectIDSchema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Description:  "The project_id you want to attach the resource to",
		Optional:     true,
		ForceNew:     true,
		Computed:     true,
		ValidateFunc: validationUUID(),
	}
}

// zoneSchema returns a standard schema for a zone
func zoneSchema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Description:  "The zone you want to attach the resource to",
		Optional:     true,
		ForceNew:     true,
		Computed:     true,
		ValidateFunc: validationZone(),
	}
}

// regionSchema returns a standard schema for a zone
func regionSchema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Description:  "The region you want to attach the resource to",
		Optional:     true,
		ForceNew:     true,
		Computed:     true,
		ValidateFunc: validationRegion(),
	}
}

// getRandomName returns a random name prefixed for terraform.
func getRandomName(prefix string) string {
	return namegenerator.GetRandomName("tf", prefix)
}

// DiffSuppressFuncForRandomName is a generic DiffSuppressFunc to avoid changing randomly generated names.
func DiffSuppressFuncForRandomName(k, old, new string, d *schema.ResourceData) bool {
	if new == "" {
		return true
	}
	return old == new
}
