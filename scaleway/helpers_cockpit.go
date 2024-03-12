package scaleway

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
	defaultCockpitTimeout = 5 * time.Minute
	pathMetricsURL        = "/api/v1/push"
	pathLogsURL           = "/loki/api/v1/push"
)

// cockpitAPI returns a new cockpit API.
func cockpitAPI(m interface{}) (*cockpit.API, error) {
	api := cockpit.NewAPI(meta.ExtractScwClient(m))

	return api, nil
}

// cockpitAPIGrafanaUserID returns a new cockpit API with the Grafana user ID and the project ID.
func cockpitAPIGrafanaUserID(m interface{}, id string) (*cockpit.API, string, uint32, error) {
	projectID, resourceIDString, err := parseCockpitID(id)
	if err != nil {
		return nil, "", 0, err
	}

	grafanaUserID, err := strconv.ParseUint(resourceIDString, 10, 32)
	if err != nil {
		return nil, "", 0, err
	}

	api, err := cockpitAPI(m)
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

func flattenCockpitEndpoints(endpoints *cockpit.CockpitEndpoints) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"metrics_url":      endpoints.MetricsURL,
			"logs_url":         endpoints.LogsURL,
			"alertmanager_url": endpoints.AlertmanagerURL,
			"grafana_url":      endpoints.GrafanaURL,
			"traces_url":       endpoints.TracesURL,
		},
	}
}

func createCockpitPushURL(endpoints *cockpit.CockpitEndpoints) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"push_metrics_url": endpoints.MetricsURL + pathMetricsURL,
			"push_logs_url":    endpoints.LogsURL + pathLogsURL,
		},
	}
}

func expandCockpitTokenScopes(raw interface{}) *cockpit.TokenScopes {
	if raw == nil || len(raw.([]interface{})) != 1 {
		return nil
	}

	rawMap := raw.([]interface{})[0].(map[string]interface{})
	return &cockpit.TokenScopes{
		QueryMetrics:      rawMap["query_metrics"].(bool),
		WriteMetrics:      rawMap["write_metrics"].(bool),
		SetupMetricsRules: rawMap["setup_metrics_rules"].(bool),
		QueryLogs:         rawMap["query_logs"].(bool),
		WriteLogs:         rawMap["write_logs"].(bool),
		SetupLogsRules:    rawMap["setup_logs_rules"].(bool),
		SetupAlerts:       rawMap["setup_alerts"].(bool),
		QueryTraces:       rawMap["query_traces"].(bool),
		WriteTraces:       rawMap["write_traces"].(bool),
	}
}

func flattenCockpitTokenScopes(scopes *cockpit.TokenScopes) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"query_metrics":       scopes.QueryMetrics,
			"write_metrics":       scopes.WriteMetrics,
			"setup_metrics_rules": scopes.SetupMetricsRules,
			"query_logs":          scopes.QueryLogs,
			"write_logs":          scopes.WriteLogs,
			"setup_logs_rules":    scopes.SetupLogsRules,
			"setup_alerts":        scopes.SetupAlerts,
			"query_traces":        scopes.QueryTraces,
			"write_traces":        scopes.WriteTraces,
		},
	}
}

func waitForCockpit(ctx context.Context, api *cockpit.API, projectID string, timeout time.Duration) (*cockpit.Cockpit, error) {
	retryInterval := defaultContainerRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	return api.WaitForCockpit(&cockpit.WaitForCockpitRequest{
		ProjectID:     projectID,
		Timeout:       scw.TimeDurationPtr(timeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))
}
