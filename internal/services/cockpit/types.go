package cockpit

import cockpit "github.com/scaleway/scaleway-sdk-go/api/cockpit/v1beta1"

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
