package cockpit

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	cockpit "github.com/scaleway/scaleway-sdk-go/api/cockpit/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

const (
	DefaultCockpitTimeout       = 5 * time.Minute
	defaultCockpitRetryInterval = 5 * time.Second
	pathMetricsURL              = "/api/v1/push"
	pathLogsURL                 = "/loki/api/v1/push"
)

// NewAPI returns a new cockpit API.
func NewAPI(m interface{}) (*cockpit.API, error) {
	api := cockpit.NewAPI(meta.ExtractScwClient(m))

	return api, nil
}

// NewAPIGrafanaUserID returns a new cockpit API with the Grafana user ID and the project ID.
func NewAPIGrafanaUserID(m interface{}, id string) (*cockpit.API, string, uint32, error) {
	projectID, resourceIDString, err := parseCockpitID(id)
	if err != nil {
		return nil, "", 0, err
	}

	grafanaUserID, err := strconv.ParseUint(resourceIDString, 10, 32)
	if err != nil {
		return nil, "", 0, err
	}

	api, err := NewAPI(m)
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

func waitForCockpit(ctx context.Context, api *cockpit.API, projectID string, timeout time.Duration) (*cockpit.Cockpit, error) {
	retryInterval := defaultCockpitRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	return api.WaitForCockpit(&cockpit.WaitForCockpitRequest{
		ProjectID:     projectID,
		Timeout:       scw.TimeDurationPtr(timeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))
}
