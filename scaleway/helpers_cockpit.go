package scaleway

import (
	"context"
	"time"

	cockpit "github.com/scaleway/scaleway-sdk-go/api/cockpit/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	defaultCockpitTimeout = 5 * time.Minute
)

// cockpitAPI returns a new container cockpit API.
func cockpitAPI(m interface{}) (*cockpit.API, error) {
	meta := m.(*Meta)
	api := cockpit.NewAPI(meta.scwClient)

	return api, nil
}

func flattenCockpitEndpoints(endpoints *cockpit.CockpitEndpoints) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"metrics_url":      endpoints.MetricsURL,
			"logs_url":         endpoints.LogsURL,
			"alertmanager_url": endpoints.AlertmanagerURL,
			"grafana_url":      endpoints.GrafanaURL,
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
		},
	}
}

func waitForCockpit(ctx context.Context, api *cockpit.API, projectID string, timeout time.Duration) (*cockpit.Cockpit, error) {
	retryInterval := defaultContainerRetryInterval
	if DefaultWaitRetryInterval != nil {
		retryInterval = *DefaultWaitRetryInterval
	}

	return api.WaitForCockpit(&cockpit.WaitForCockpitRequest{
		ProjectID:     projectID,
		Timeout:       scw.TimeDurationPtr(timeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))
}
