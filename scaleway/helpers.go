package scaleway

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	api "github.com/nicolai86/scaleway-sdk"
	"github.com/scaleway/scaleway-sdk-go/namegenerator"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"golang.org/x/xerrors"
)

// RegionalID represents an ID that is linked with a region, eg fr-par/11111111-1111-1111-1111-111111111111
type RegionalID struct {
	ID     string
	Region scw.Region
}

func (z RegionalID) String() string {
	return fmt.Sprintf("%s/%s", z.Region, z.ID)
}

func newRegionalID(region scw.Region, id string) RegionalID {
	return RegionalID{
		ID:     id,
		Region: region,
	}
}

func expandRegionalID(id interface{}) RegionalID {
	regionalID := RegionalID{}
	tab := strings.SplitN(id.(string), "/", -1)
	if len(tab) != 2 {
		regionalID.ID = id.(string)
	} else {
		region, _ := scw.ParseRegion(tab[0])
		regionalID.ID = tab[1]
		regionalID.Region = region
	}

	return regionalID
}

// ZonedID represents an ID that is linked with a zone, eg fr-par-1/11111111-1111-1111-1111-111111111111
type ZonedID struct {
	ID   string
	Zone scw.Zone
}

func (z ZonedID) String() string {
	return fmt.Sprintf("%s/%s", z.Zone, z.ID)
}

func newZonedID(zone scw.Zone, id string) ZonedID {
	return ZonedID{
		ID:   id,
		Zone: zone,
	}
}

func expandZonedID(id interface{}) ZonedID {
	zonedID := ZonedID{}
	tab := strings.SplitN(id.(string), "/", -1)
	if len(tab) != 2 {
		zonedID.ID = id.(string)
	} else {
		zone, _ := scw.ParseZone(tab[0])
		zonedID.ID = tab[1]
		zonedID.Zone = zone
	}

	return zonedID
}

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
		isKnown = isKnown || strings.EqualFold(knownType, requestedType)
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

func waitForServerShutdown(scaleway *api.API, serverID string) error {
	return waitForServerState(scaleway, serverID, "stopped", []string{"stopped", "stopping"})
}

func waitForServerStartup(scaleway *api.API, serverID string) error {
	return waitForServerState(scaleway, serverID, "running", []string{"running", "starting"})
}

func waitForServerState(scaleway *api.API, serverID, targetState string, pendingStates []string) error {
	wg := waitForServerLock(serverID)
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

func waitForServerLock(serverID string) *sync.WaitGroup {
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
	wg := waitForServerLock(server.Identifier)
	wg.Wait()

	_, err := scaleway.PostServerAction(server.Identifier, "poweron")

	if err != nil {
		return err
	}

	return waitForServerStartup(scaleway, server.Identifier)
}

func stopServer(scaleway *api.API, server *api.Server) error {
	wg := waitForServerLock(server.Identifier)
	wg.Wait()

	_, err := scaleway.PostServerAction(server.Identifier, "poweroff")

	if err != nil {
		return err
	}
	return waitForServerShutdown(scaleway, server.Identifier)
}

// deleteRunningServer terminates the server and waits until it is removed.
func deleteRunningServer(scaleway *api.API, server *api.Server) error {
	wg := waitForServerLock(server.Identifier)
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
	wg := waitForServerLock(serverID)
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
	tab := strings.SplitN(localizedID, "/", -1)
	if len(tab) != 2 {
		return "", localizedID, fmt.Errorf("cant parse localized id: %s", localizedID)
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

// deprecated and should not be used
// newZonedIdFromRegion constructs a unique identifier based on resource region and id
// but returns a zoned ID with the first zone in the region, i.e. adding `-1` to the region
// TODO this function is a quick fix
func newZonedIdFromRegion(region scw.Region, id string) string {
	return fmt.Sprintf("%s-1/%s", region, id)
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
var ErrZoneNotFound = fmt.Errorf("could not detect zone. Scaleway uses regions and zones. For more information, refer to https://www.terraform.io/docs/providers/scaleway/guides/regions_and_zones.html")

// extractZone will try to guess the zone from the following:
//  - zone field of the resource data
//  - default zone from config
func extractZone(d terraformResourceData, meta *Meta) (scw.Zone, error) {

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

// extractRegion will try to guess the region from the following:
//  - region field of the resource data
//  - default region from config
func extractRegion(d terraformResourceData, meta *Meta) (scw.Region, error) {

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

// ErrOrganizationIDNotFound is returned when no organization_id can be detected
var ErrOrganizationIDNotFound = fmt.Errorf("could not detect organization_id")

// organizationID will try to guess the organization_id from the following:
//  - organization_id field of the resource data
//  - default organization_id from config
func organizationID(d terraformResourceData, meta *Meta) (string, error) {

	organizationID, exist := d.GetOkExists("organization_id")
	if exist {
		return organizationID.(string), nil
	}

	organizationID, exist = meta.scwClient.GetDefaultOrganizationID()
	if exist {
		return organizationID.(string), nil
	}

	return "", ErrOrganizationIDNotFound
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
	notFoundError := &scw.ResourceNotFoundError{}
	return isHTTPCodeError(err, http.StatusNotFound) || xerrors.As(err, &notFoundError)
}

// is403Error returns true if err is an HTTP 403 error
func is403Error(err error) bool {
	permissionsDeniedError := &scw.PermissionsDeniedError{}
	return isHTTPCodeError(err, http.StatusForbidden) || xerrors.As(err, &permissionsDeniedError)
}

// organizationIDSchema returns a standard schema for a organization_id
func organizationIDSchema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Description:  "The organization_id you want to attach the resource to",
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

// newRandomName returns a random name prefixed for terraform.
func newRandomName(prefix string) string {
	return namegenerator.GetRandomName("tf", prefix)
}

const gb uint64 = 1000 * 1000 * 1000

var UUIDRegex = regexp.MustCompile(`[0-9a-fA-F]{8}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{12}`)

// isUUID returns true if the given string have an UUID format.
func isUUID(s string) bool {
	return UUIDRegex.MatchString(s)
}

// newTemplateFunc takes a go template string and returns a function that can be called to execute template.
func newTemplateFunc(tplStr string) func(data interface{}) string {
	t := template.Must(template.New("tpl").Parse(tplStr))
	return func(tplParams interface{}) string {
		buffer := bytes.Buffer{}
		err := t.Execute(&buffer, tplParams)
		if err != nil {
			panic(err)
		}
		return buffer.String()
	}
}

// testAccGetResourceAttr can be used in acceptance tests to extract value from state and store it in dest
func testAccGetResourceAttr(resourceName string, attrName string, dest *string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		r, exist := state.RootModule().Resources[resourceName]
		if !exist {
			return fmt.Errorf("unknown ressource %s", resourceName)
		}

		a, exist := r.Primary.Attributes[attrName]
		if !exist {
			return fmt.Errorf("unknown ressource %s", resourceName)
		}

		*dest = a
		return nil
	}
}

func flattenTime(date *time.Time) interface{} {
	if date != nil {
		return date.Format(time.RFC3339)
	}
	return ""
}

func flattenDuration(duration *time.Duration) interface{} {
	if duration != nil {
		return duration.String()
	}
	return ""
}

func expandDuration(data interface{}) *time.Duration {
	if data == nil || data == "" {
		return nil
	}
	d, err := time.ParseDuration(data.(string))
	if err != nil {
		// We panic as this should never happend. Data from state should be validate using a validate func
		panic(err)
	}
	return &d
}

func expandOrGenerateString(data interface{}, prefix string) string {
	if data == nil || data == "" {
		return newRandomName(prefix)
	}
	return data.(string)
}

func expandStringWithDefault(data interface{}, defaultValue string) string {
	if data == nil || data.(string) == "" {
		return defaultValue
	}
	return data.(string)
}

func expandStrings(data interface{}) []string {
	if data == nil {
		return nil
	}
	stringSlice := []string(nil)
	for _, s := range data.([]interface{}) {
		stringSlice = append(stringSlice, s.(string))
	}
	return stringSlice
}

func expandSliceStringPtr(data interface{}) []*string {
	if data == nil {
		return nil
	}
	stringSlice := []*string(nil)
	for _, s := range data.([]interface{}) {
		stringSlice = append(stringSlice, expandStringPtr(s))
	}
	return stringSlice
}

func flattenIPPtr(ip *net.IP) interface{} {
	if ip == nil {
		return ""
	}
	return ip.String()
}

func flattenStringPtr(s *string) interface{} {
	if s == nil {
		return ""
	}
	return *s
}

func flattenSliceStringPtr(s []*string) interface{} {
	res := make([]interface{}, 0, len(s))
	for _, strPtr := range s {
		res = append(res, flattenStringPtr(strPtr))
	}
	return res
}

func expandStringPtr(data interface{}) *string {
	if data == nil || data == "" {
		return nil
	}
	return scw.StringPtr(data.(string))
}

func flattenInt32Ptr(i *int32) interface{} {
	if i == nil {
		return 0
	}
	return *i
}

func expandInt32Ptr(data interface{}) *int32 {
	if data == nil || data == "" {
		return nil
	}
	return scw.Int32Ptr(int32(data.(int)))
}

func expandIPNet(raw string) scw.IPNet {
	if raw == "" {
		return scw.IPNet{}
	}
	var ipNet scw.IPNet
	raw = `"` + raw + `"`
	err := json.Unmarshal([]byte(raw), &ipNet)
	if err != nil {
		// We panic as this should never happen. Data from state should be validate using a validate func
		panic(fmt.Errorf("%s could not be marshaled: %v", raw, err))
	}

	return ipNet
}

func flattenIpNet(ipNet scw.IPNet) string {
	raw, err := json.Marshal(ipNet)
	if err != nil {
		// We panic as this should never happen.
		panic(err)
	}
	return string(raw[1 : len(raw)-1])
}

func validateDuration() schema.SchemaValidateFunc {
	return func(i interface{}, s string) (strings []string, errors []error) {
		str, isStr := i.(string)
		if !isStr {
			return nil, []error{fmt.Errorf("%v is not a string", i)}
		}
		_, err := time.ParseDuration(str)
		if err != nil {
			return nil, []error{fmt.Errorf("cannot parse duration for value %s", str)}
		}
		return nil, nil
	}
}

func validateHour() schema.SchemaValidateFunc {
	return func(i interface{}, s string) (strings []string, errors []error) {
		integer, isInteger := i.(int)
		if !isInteger {
			return nil, []error{fmt.Errorf("%v is not an int", i)}
		}
		if integer < 0 || integer > 23 {
			return nil, []error{fmt.Errorf("int is outside range 0-23 for value %d", integer)}
		}
		return nil, nil
	}
}

func diffSuppressFuncDuration(k, old, new string, d *schema.ResourceData) bool {
	if old == new {
		return true
	}
	d1, err1 := time.ParseDuration(old)
	d2, err2 := time.ParseDuration(new)
	if err1 != nil || err2 != nil {
		return false
	}
	return d1 == d2
}

func diffSuppressFuncIgnoreCase(k, old, new string, d *schema.ResourceData) bool {
	return strings.EqualFold(old, new)
}

func diffSuppressFuncIgnoreCaseAndHyphen(k, old, new string, d *schema.ResourceData) bool {
	return strings.Replace(strings.ToLower(old), "-", "_", -1) == strings.Replace(strings.ToLower(new), "-", "_", -1)
}

// diffSuppressFuncLocality is a SuppressDiffFunc to remove the locality from an ID when checking diff.
// e.g. 2c1a1716-5570-4668-a50a-860c90beabf6 == fr-par/2c1a1716-5570-4668-a50a-860c90beabf6
func diffSuppressFuncLocality(k, old, new string, d *schema.ResourceData) bool {
	return expandID(old) == expandID(new)
}
