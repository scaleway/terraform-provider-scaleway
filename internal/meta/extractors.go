package meta

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
)

// terraformResourceData is an interface for *schema.ResourceData. (used for mock)
type terraformResourceData interface {
	HasChange(string) bool
	GetOk(string) (any, bool)
	Get(string) any
	Id() string
}

// ExtractZone will try to guess the zone from the following:
//   - zone field of the resource data
//   - default zone from config
func ExtractZone(d terraformResourceData, m any) (scw.Zone, error) {
	rawZone, exist := d.GetOk("zone")
	if exist {
		return scw.ParseZone(rawZone.(string))
	}

	zone, exist := m.(*Meta).ScwClient().GetDefaultZone()
	if exist {
		return zone, nil
	}

	return "", zonal.ErrZoneNotFound
}

// ExtractRegion will try to guess the region from the following:
//   - region field of the resource data
//   - default region from config
func ExtractRegion(d terraformResourceData, m any) (scw.Region, error) {
	rawRegion, exist := d.GetOk("region")
	if exist {
		return scw.ParseRegion(rawRegion.(string))
	}

	region, exist := m.(*Meta).ScwClient().GetDefaultRegion()
	if exist {
		return region, nil
	}

	return "", regional.ErrRegionNotFound
}

// ExtractRegionWithDefault will try to guess the region from the following:
//   - region field of the resource data
//   - default region given in argument
//   - default region from config
func ExtractRegionWithDefault(d terraformResourceData, m any, defaultRegion scw.Region) (scw.Region, error) {
	rawRegion, exist := d.GetOk("region")
	if exist {
		return scw.ParseRegion(rawRegion.(string))
	}

	if defaultRegion != "" {
		return defaultRegion, nil
	}

	region, exist := m.(*Meta).ScwClient().GetDefaultRegion()
	if exist {
		return region, nil
	}

	return "", regional.ErrRegionNotFound
}

// ExtractProjectID will try to guess the project id from the following:
//   - project_id field of the resource data
//   - default project id from config
func ExtractProjectID(d terraformResourceData, m any) (projectID string, isDefault bool, err error) {
	rawProjectID, exist := d.GetOk("project_id")
	if exist {
		return rawProjectID.(string), false, nil
	}

	defaultProjectID, exist := m.(*Meta).ScwClient().GetDefaultProjectID()
	if exist {
		return defaultProjectID, true, nil
	}

	return "", false, ErrProjectIDNotFound
}

func ExtractScwClient(m any) *scw.Client {
	return m.(*Meta).ScwClient()
}

func ExtractHTTPClient(m any) *http.Client {
	return m.(*Meta).HTTPClient()
}

func getKeyInRawConfigMap(rawConfig map[string]cty.Value, key string, ty cty.Type) (any, bool) {
	if key == "" {
		return rawConfig, false
	}
	// We split the key into its elements
	keys := strings.Split(key, ".")

	// We look at the first element's type
	if value, ok := rawConfig[keys[0]]; ok {
		switch {
		case value.Type().IsListType():
			// If it's a list and the second element of the key is an index, we look for the value in the list at the given index
			if index, err := strconv.Atoi(keys[1]); err == nil {
				return getKeyInRawConfigMap(value.AsValueSlice()[index].AsValueMap(), strings.Join(keys[2:], ""), ty)
			}
			// If it's a list and the second element of the key is '#', we look for the value in the list's first element
			return getKeyInRawConfigMap(value.AsValueSlice()[0].AsValueMap(), strings.Join(keys[2:], ""), ty)

		case value.Type().IsMapType():
			// If it's a map, we look for the value in the map
			return getKeyInRawConfigMap(value.AsValueMap(), strings.Join(keys[1:], ""), ty)

		case value.Type().IsPrimitiveType():
			// If it's a primitive type (bool, string, number), we convert the value to the expected type given as parameter before returning it
			switch ty {
			case cty.String:
				if value.IsNull() {
					return nil, false
				}

				return value.AsString(), true
			case cty.Bool:
				if value.IsNull() {
					return false, false
				}

				if value.True() {
					return true, true
				}

				return false, true
			case cty.Number:
				if value.IsNull() {
					return nil, false
				}

				valueInt, _ := value.AsBigFloat().Int64()

				return valueInt, true
			}
		}
	}

	return nil, false
}

// GetRawConfigForKey returns the value for a specific key in the user's raw configuration, which can be useful on resources' update
// The value for the key to look for must be a primitive type (bool, string, number) and the expected type of the value should be passed as the ty parameter
func GetRawConfigForKey(d *schema.ResourceData, key string, ty cty.Type) (any, bool) {
	rawConfig := d.GetRawConfig()
	if rawConfig.IsNull() {
		return nil, false
	}

	return getKeyInRawConfigMap(rawConfig.AsValueMap(), key, ty)
}
