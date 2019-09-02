package scaleway

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	api "github.com/nicolai86/scaleway-sdk"
	"github.com/scaleway/scaleway-sdk-go/namegenerator"
	"github.com/scaleway/scaleway-sdk-go/scw"
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
func parseZonedID(zonedID string) (zone scw.Zone, id string, err error) {

	locality, id, err := parseLocalizedID(zonedID)
	if err != nil {
		return
	}

	zone, err = scw.ParseZone(locality)
	return
}

// expandID returns the id whether it is a localizedID or a raw ID.
func expandID(id interface{}) string {
	_, ID, err := parseLocalizedID(id.(string))
	if err != nil {
		return id.(string)
	}
	return ID
}

// parseRegionalID parses a regionalID and extracts the resource region and id.
func parseRegionalID(regionalID string) (region scw.Region, id string, err error) {
	locality, id, err := parseLocalizedID(regionalID)
	if err != nil {
		return
	}

	region, err = scw.ParseRegion(locality)
	return
}

// newZonedId constructs a unique identifier based on resource zone and id
func newZonedId(zone scw.Zone, id string) string {
	return fmt.Sprintf("%s/%s", zone, id)
}

// newRegionalId constructs a unique identifier based on resource region and id
func newRegionalId(region scw.Region, id string) string {
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

// ErrZoneNotFound is returned when no zone can be detected
var ErrZoneNotFound = fmt.Errorf("could not detect zone")

// getZone will try to guess the zone from the following:
//  - zone field of the resource data
//  - default zone from config
func getZone(d terraformResourceData, meta *Meta) (scw.Zone, error) {

	rawZone, exist := d.GetOkExists("zone")
	if exist {
		return scw.ParseZone(rawZone.(string))
	}

	zone, exist := meta.scwClient.GetDefaultZone()
	if exist {
		return zone, nil
	}

	return scw.Zone(""), ErrZoneNotFound
}

// ErrRegionNotFound is returned when no region can be detected
var ErrRegionNotFound = fmt.Errorf("could not detect region")

// getRegion will try to guess the region from the following:
//  - region field of the resource data
//  - default region from config
func getRegion(d terraformResourceData, meta *Meta) (scw.Region, error) {

	rawRegion, exist := d.GetOkExists("region")
	if exist {
		return scw.ParseRegion(rawRegion.(string))
	}

	region, exist := meta.scwClient.GetDefaultRegion()
	if exist {
		return region, nil
	}

	return scw.Region(""), ErrRegionNotFound
}

// ErrProjectIdNotFound is returned when no project_id can be detected
var ErrProjectIdNotFound = fmt.Errorf("could not detect project_id")

// getProjectId will try to guess the project_id from the following:
//  - project_id field of the resource data
//  - default project_id from config
func getProjectId(d terraformResourceData, meta *Meta) (string, error) {

	projectID, exist := d.GetOkExists("project_id")
	if exist {
		return projectID.(string), nil
	}

	projectID, exist = meta.scwClient.GetDefaultProjectID()
	if exist {
		return projectID.(string), nil
	}

	return "", ErrProjectIdNotFound
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

const gb uint64 = 1000 * 1000 * 1000

// suppressLocality is a SuppressDiffFunc to remove the locality from an ID when checking diff.
// e.g. 2c1a1716-5570-4668-a50a-860c90beabf6 == fr-par/2c1a1716-5570-4668-a50a-860c90beabf6
func suppressLocality(k, old, new string, d *schema.ResourceData) bool {
	return expandID(old) == expandID(new)
}

// isResourceTimeoutError returns true when the given error is a timeout error returned by
// terraform's Retry helper.
func isResourceTimeoutError(err error) bool {
	timeoutErr, ok := err.(*resource.TimeoutError)
	return ok && timeoutErr.LastError == nil
}

// isSDKResponseError returns true when the given http status and the message match
// with the scw.ResponseError status and message.
func isSDKResponseError(err error, status int, message string) bool {
	responseError, ok := err.(*scw.ResponseError)
	if !ok {
		return false
	}

	return responseError.StatusCode == status && responseError.Message == message
}

// isSDKError returns true when the SdkError error message matches with the given message.
func isSDKError(err error, expectedMessage string) bool {

	responseError, ok := err.(scw.SdkError)
	if !ok {
		return false
	}
	actualMessage := responseError.Error()[17:] // remove "scaleway-sdk-go: "
	if actualMessage == expectedMessage {
		return true
	}

	regexp, err := regexp.Compile(expectedMessage)
	if err != nil {
		return false
	}

	return regexp.MatchString(actualMessage)
}

// isUUID returns true if the given string have an UUID format.
func isUUID(s string) bool {
	matched, _ := regexp.MatchString(`[0-9a-fA-F]{8}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{12}`, s)
	return matched
}
