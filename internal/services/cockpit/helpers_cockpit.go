package cockpit

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/scaleway-sdk-go/validation"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

const (
	DefaultCockpitTimeout = 5 * time.Minute
	pathMetricsURL        = "/api/v1/push"
	pathLogsURL           = "/loki/api/v1/push"
	pathTracesURL         = "/otlp/v1/traces"
)

// NewGlobalAPI returns a new global cockpit API.
func NewGlobalAPI(m interface{}) (*cockpit.GlobalAPI, error) {
	api := cockpit.NewGlobalAPI(meta.ExtractScwClient(m))

	return api, nil
}

func cockpitAPIWithRegion(d *schema.ResourceData, m interface{}) (*cockpit.RegionalAPI, scw.Region, error) {
	api := cockpit.NewRegionalAPI(meta.ExtractScwClient(m))

	region, err := meta.ExtractRegion(d, m)
	if err != nil {
		return nil, "", err
	}
	return api, region, err
}

func NewAPIWithRegionAndID(m interface{}, id string) (*cockpit.RegionalAPI, scw.Region, string, error) {
	api := cockpit.NewRegionalAPI(meta.ExtractScwClient(m))

	region, id, err := regional.ParseID(id)
	if err != nil {
		return nil, "", "", err
	}
	return api, region, id, nil
}

// NewAPIGrafanaUserID returns a new cockpit API with the Grafana user ID and the project ID.
func NewAPIGrafanaUserID(m interface{}, id string) (*cockpit.GlobalAPI, string, uint32, error) {
	projectID, resourceIDString, err := parseCockpitID(id)
	if err != nil {
		return nil, "", 0, err
	}

	grafanaUserID, err := strconv.ParseUint(resourceIDString, 10, 32)
	if err != nil {
		return nil, "", 0, err
	}

	api, err := NewGlobalAPI(m)
	if err != nil {
		return nil, "", 0, err
	}

	return api, projectID, uint32(grafanaUserID), nil
}

// cockpitIDWithProjectID returns a cockpit ID with a project ID.
func cockpitIDWithProjectID(projectID string, id string) string {
	return projectID + "/" + id
}

// parseCockpitID returns the project ID and the cockpit ID from a combined ID.
func parseCockpitID(id string) (projectID string, cockpitID string, err error) {
	parts := strings.Split(id, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid cockpit ID: %s", id)
	}
	return parts[0], parts[1], nil
}

func cockpitTokenUpgradeV1SchemaType() cty.Type {
	return cty.Object(map[string]cty.Type{
		"id": cty.String,
	})
}

func cockpitTokenV1UpgradeFunc(_ context.Context, rawState map[string]interface{}, _ interface{}) (map[string]interface{}, error) {
	defaultRegion := scw.RegionFrPar // Default to the 'fr-par' region as all tokens created with the v1beta1 API were implicitly set to this region

	if _, ok := rawState["region"]; !ok {
		rawState["region"] = defaultRegion.String()
	}

	if id, ok := rawState["id"].(string); ok && validation.IsUUID(id) {
		locality, ID, _ := regional.ParseID(id)
		if locality == "" {
			rawState["id"] = regional.NewIDString(defaultRegion, ID)
		}
	} else {
		return nil, fmt.Errorf("upgrade: expected 'id' to be a string, got %T", rawState["id"])
	}

	return rawState, nil
}
