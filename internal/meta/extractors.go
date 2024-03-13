package meta

import (
	"net/http"

	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
)

// terraformResourceData is an interface for *schema.ResourceData. (used for mock)
type terraformResourceData interface {
	HasChange(string) bool
	GetOk(string) (interface{}, bool)
	Get(string) interface{}
	Id() string
}

// ExtractZone will try to guess the zone from the following:
//   - zone field of the resource data
//   - default zone from config
func ExtractZone(d terraformResourceData, m interface{}) (scw.Zone, error) {
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
func ExtractRegion(d terraformResourceData, m interface{}) (scw.Region, error) {
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
func ExtractRegionWithDefault(d terraformResourceData, m interface{}, defaultRegion scw.Region) (scw.Region, error) {
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
func ExtractProjectID(d terraformResourceData, m interface{}) (projectID string, isDefault bool, err error) {
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

func ExtractScwClient(m interface{}) *scw.Client {
	return m.(*Meta).ScwClient()
}

func ExtractHTTPClient(m interface{}) *http.Client {
	return m.(*Meta).HTTPClient()
}
