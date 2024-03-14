package scaleway

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

// Service information constants
const (
	ServiceName = "scw"       // Name of service.
	EndpointsID = ServiceName // ID to look up a service endpoint with.
)

// organizationIDSchema returns a standard schema for a organization_id
func organizationIDSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Description: "The organization_id you want to attach the resource to",
		Computed:    true,
	}
}

func organizationIDOptionalSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
		Description: "ID of organization the resource is associated to.",
	}
}

// projectIDSchema returns a standard schema for a project_id
func projectIDSchema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Description:  "The project_id you want to attach the resource to",
		Optional:     true,
		ForceNew:     true,
		Computed:     true,
		ValidateFunc: verify.IsUUID(),
	}
}

const gb uint64 = 1000 * 1000 * 1000

func validateDuration() schema.SchemaValidateFunc {
	return func(i interface{}, _ string) (strings []string, errors []error) {
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

func diffSuppressFuncDuration(_, oldValue, newValue string, _ *schema.ResourceData) bool {
	if oldValue == newValue {
		return true
	}
	d1, err1 := time.ParseDuration(oldValue)
	d2, err2 := time.ParseDuration(newValue)
	if err1 != nil || err2 != nil {
		return false
	}
	return d1 == d2
}

func diffSuppressFuncTimeRFC3339(_, oldValue, newValue string, _ *schema.ResourceData) bool {
	if oldValue == newValue {
		return true
	}
	t1, err1 := time.Parse(time.RFC3339, oldValue)
	t2, err2 := time.Parse(time.RFC3339, newValue)
	if err1 != nil || err2 != nil {
		return false
	}
	return t1.Equal(t2)
}

func diffSuppressFuncIgnoreCase(_, oldValue, newValue string, _ *schema.ResourceData) bool {
	return strings.EqualFold(oldValue, newValue)
}

func diffSuppressFuncIgnoreCaseAndHyphen(_, oldValue, newValue string, _ *schema.ResourceData) bool {
	return strings.ReplaceAll(strings.ToLower(oldValue), "-", "_") == strings.ReplaceAll(strings.ToLower(newValue), "-", "_")
}

// diffSuppressFuncLocality is a SuppressDiffFunc to remove the locality from an ID when checking diff.
// e.g. 2c1a1716-5570-4668-a50a-860c90beabf6 == fr-par-1/2c1a1716-5570-4668-a50a-860c90beabf6
func diffSuppressFuncLocality(_, oldValue, newValue string, _ *schema.ResourceData) bool {
	return locality.ExpandID(oldValue) == locality.ExpandID(newValue)
}

// diffSuppressFuncOrderDiff suppresses diffs for TypeList attributes when the only change is the order of elements.
// https://github.com/hashicorp/terraform-plugin-sdk/issues/477#issuecomment-1238807249
func diffSuppressFuncOrderDiff(k, _, _ string, d *schema.ResourceData) bool {
	baseKey := extractBaseKey(k)
	oldList, newList := getStringListsFromState(baseKey, d)

	return compareStringListsIgnoringOrder(oldList, newList)
}

func extractBaseKey(k string) string {
	lastDotIndex := strings.LastIndex(k, ".")
	if lastDotIndex != -1 {
		return k[:lastDotIndex]
	}

	return k
}

func getStringListsFromState(key string, d *schema.ResourceData) ([]string, []string) {
	oldList, newList := d.GetChange(key)

	oldListStr := make([]string, len(oldList.([]interface{})))
	newListStr := make([]string, len(newList.([]interface{})))

	for i, v := range oldList.([]interface{}) {
		oldListStr[i] = fmt.Sprint(v)
	}
	for i, v := range newList.([]interface{}) {
		newListStr[i] = fmt.Sprint(v)
	}

	return oldListStr, newListStr
}

func compareStringListsIgnoringOrder(oldListStr, newListStr []string) bool {
	if len(oldListStr) != len(newListStr) {
		return false // different lengths means there's definitely a change
	}

	sort.Strings(oldListStr)
	sort.Strings(newListStr)

	return reflect.DeepEqual(oldListStr, newListStr)
}

// TimedOut returns true if the error represents a "wait timed out" condition.
// Specifically, TimedOut returns true if the error matches all these conditions:
//   - err is of type resource.TimeoutError
//   - TimeoutError.LastError is nil
func TimedOut(err error) bool {
	// This explicitly does *not* match wrapped TimeoutErrors
	timeoutErr, ok := err.(*resource.TimeoutError) //nolint:errorlint // Explicitly does *not* match wrapped TimeoutErrors
	return ok && timeoutErr.LastError == nil
}

func errorCheck(err error, message string) bool {
	return strings.Contains(err.Error(), message)
}

// ErrCodeEquals returns true if the error matches all these conditions:
//   - err is of type scw.Error
//   - Error.Error() equals one of the passed codes
func ErrCodeEquals(err error, codes ...string) bool {
	var scwErr scw.SdkError
	if errors.As(err, &scwErr) {
		for _, code := range codes {
			if scwErr.Error() == code {
				return true
			}
		}
	}
	return false
}

func getBool(d *schema.ResourceData, key string) interface{} {
	val, ok := d.GetOkExists(key)
	if !ok {
		return nil
	}
	return val
}

// validateDate will validate that field is a valid ISO 8601
// It is the same as RFC3339
func validateDate() schema.SchemaValidateDiagFunc {
	return func(i interface{}, _ cty.Path) diag.Diagnostics {
		date, isStr := i.(string)
		if !isStr {
			return diag.Errorf("%v is not a string", date)
		}
		_, err := time.Parse(time.RFC3339, date)
		if err != nil {
			return diag.FromErr(err)
		}
		return nil
	}
}

type ServiceErrorCheckFunc func(*testing.T) resource.ErrorCheckFunc

var serviceErrorCheckFunc map[string]ServiceErrorCheckFunc

func ErrorCheck(t *testing.T, endpointIDs ...string) resource.ErrorCheckFunc {
	t.Helper()
	return func(err error) error {
		if err == nil {
			return nil
		}

		for _, endpointID := range endpointIDs {
			if f, ok := serviceErrorCheckFunc[endpointID]; ok {
				ef := f(t)
				err = ef(err)
			}

			if err == nil {
				break
			}
		}

		return err
	}
}

func validateMapKeyLowerCase() schema.SchemaValidateDiagFunc {
	return func(i interface{}, _ cty.Path) diag.Diagnostics {
		m := types.ExpandMapStringStringPtr(i)
		for k := range m {
			if strings.ToLower(k) != k {
				return diag.Diagnostics{diag.Diagnostic{
					Severity:      diag.Error,
					AttributePath: cty.IndexStringPath(k),
					Summary:       "Invalid map content",
					Detail:        fmt.Sprintf("key (%s) should be lowercase", k),
				}}
			}
		}
		return nil
	}
}

// expandListKeys return the list of keys for an attribute in a list
// example for private-networks.#.id in a list of size 2
// will return private-networks.0.id and private-networks.1.id
// additional_volume_ids.#
// will return additional_volume_ids.0 and additional_volume_ids.1
func expandListKeys(key string, diff *schema.ResourceDiff) []string {
	addr := strings.Split(key, ".")
	// index of # in the addr
	index := 0

	for i := range addr {
		if addr[i] == "#" {
			index = i
		}
	}

	// get attribute.#
	listKey := key[:strings.Index(key, "#")+1]
	listLength := diff.Get(listKey).(int)

	keys := make([]string, 0, listLength)

	for i := 0; i < listLength; i++ {
		addr[index] = strconv.FormatInt(int64(i), 10)
		keys = append(keys, strings.Join(addr, "."))
	}

	return keys
}

// getLocality find the locality of a resource
// Will try to get the zone if available then use region
// Will also use default zone or region if available
func getLocality(diff *schema.ResourceDiff, m interface{}) string {
	var loc string

	rawStateType := diff.GetRawState().Type()

	if rawStateType.HasAttribute("zone") {
		zone, _ := meta.ExtractZone(diff, m)
		loc = zone.String()
	} else if rawStateType.HasAttribute("region") {
		region, _ := meta.ExtractRegion(diff, m)
		loc = region.String()
	}
	return loc
}

// CustomizeDiffLocalityCheck create a function that will validate locality IDs stored in given keys
// This locality IDs should have the same locality as the resource
// It will search for zone or region in resource.
// Should not be used on computed keys, if a computed key is going to change on zone/region change
// this function will still block the terraform plan
func CustomizeDiffLocalityCheck(keys ...string) schema.CustomizeDiffFunc {
	return func(_ context.Context, diff *schema.ResourceDiff, m interface{}) error {
		l := getLocality(diff, m)

		if l == "" {
			return errors.New("missing locality zone or region to check IDs")
		}

		for _, key := range keys {
			// Handle values in lists
			if strings.Contains(key, "#") {
				listKeys := expandListKeys(key, diff)

				for _, listKey := range listKeys {
					IDLocality, _, err := locality.ParseLocalizedID(diff.Get(listKey).(string))
					if err == nil && !locality.CompareLocalities(IDLocality, l) {
						return fmt.Errorf("given %s %s has different locality than the resource %q", listKey, diff.Get(listKey), l)
					}
				}
			} else {
				IDLocality, _, err := locality.ParseLocalizedID(diff.Get(key).(string))
				if err == nil && !locality.CompareLocalities(IDLocality, l) {
					return fmt.Errorf("given %s %s has different locality than the resource %q", key, diff.Get(key), l)
				}
			}
		}
		return nil
	}
}

type TooManyResultsError struct {
	Count       int
	LastRequest interface{}
}

func (e *TooManyResultsError) Error() string {
	return fmt.Sprintf("too many results: wanted 1, got %d", e.Count)
}

func (e *TooManyResultsError) Is(err error) bool {
	_, ok := err.(*TooManyResultsError) //nolint:errorlint // Explicitly does *not* match down the error tree
	return ok
}

func (e *TooManyResultsError) As(target interface{}) bool {
	t, ok := target.(**retry.NotFoundError)
	if !ok {
		return false
	}

	*t = &retry.NotFoundError{
		Message:     e.Error(),
		LastRequest: e.LastRequest,
	}

	return true
}

var ErrTooManyResults = &TooManyResultsError{}

// SingularDataSourceFindError returns a standard error message for a singular data source's non-nil resource find error.
func SingularDataSourceFindError(resourceType string, err error) error {
	if NotFound(err) {
		if errors.Is(err, &TooManyResultsError{}) {
			return fmt.Errorf("multiple %[1]ss matched; use additional constraints to reduce matches to a single %[1]s", resourceType)
		}

		return fmt.Errorf("no matching %[1]s found", resourceType)
	}

	return fmt.Errorf("reading %s: %w", resourceType, err)
}

// NotFound returns true if the error represents a "resource not found" condition.
// Specifically, NotFound returns true if the error or a wrapped error is of type
// retry.NotFoundError.
func NotFound(err error) bool {
	var e *retry.NotFoundError // nosemgrep:ci.is-not-found-error
	return errors.As(err, &e)
}

func sliceContainsString(slice []string, str string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
}

// findExact finds the first element in 'slice' matching the condition defined by 'finder'.
// It returns the first matching element and an error if either no match is found or multiple matches are found.
func findExact[T any](slice []T, finder func(T) bool, searchName string) (T, error) { //nolint:ireturn
	var found T
	var foundFlag bool

	for _, elem := range slice {
		if finder(elem) {
			if foundFlag {
				// More than one element found with the same search name
				var zero T
				return zero, fmt.Errorf("multiple elements found with the name %s", searchName)
			}
			found = elem
			foundFlag = true
		}
	}

	if !foundFlag {
		var zero T
		return zero, fmt.Errorf("no element found with the name %s", searchName)
	}

	return found, nil
}
