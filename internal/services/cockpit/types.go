package cockpit

import (
	"github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	cockpitv1beta1 "github.com/scaleway/scaleway-sdk-go/api/cockpit/v1beta1"
)

var scopeMapping = map[string]cockpit.TokenScope{
	"query_metrics":       cockpit.TokenScopeReadOnlyMetrics,
	"write_metrics":       cockpit.TokenScopeWriteOnlyMetrics,
	"setup_metrics_rules": cockpit.TokenScopeFullAccessMetricsRules,
	"query_logs":          cockpit.TokenScopeReadOnlyLogs,
	"write_logs":          cockpit.TokenScopeWriteOnlyLogs,
	"setup_logs_rules":    cockpit.TokenScopeFullAccessLogsRules,
	"setup_alerts":        cockpit.TokenScopeFullAccessAlertManager,
	"query_traces":        cockpit.TokenScopeReadOnlyTraces,
	"write_traces":        cockpit.TokenScopeWriteOnlyTraces,
}

func flattenCockpitEndpoints(endpoints *cockpitv1beta1.CockpitEndpoints) []map[string]interface{} {
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

func createCockpitPushURL(endpoints *cockpitv1beta1.CockpitEndpoints) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"push_metrics_url": endpoints.MetricsURL + pathMetricsURL,
			"push_logs_url":    endpoints.LogsURL + pathLogsURL,
		},
	}
}

func expandCockpitTokenScopes(raw interface{}) []cockpit.TokenScope {
	var expandedScopes []cockpit.TokenScope

	scopesList, ok := raw.([]interface{})
	if !ok || len(scopesList) == 0 {
		return expandedScopes
	}

	scopesMap, ok := scopesList[0].(map[string]interface{})
	if !ok {
		return expandedScopes
	}

	for key, tokenScope := range scopeMapping {
		if value, ok := scopesMap[key].(bool); ok && value {
			expandedScopes = append(expandedScopes, tokenScope)
		}
	}

	return expandedScopes
}

func flattenCockpitTokenScopes(scopes []cockpit.TokenScope) []interface{} {
	result := map[string]interface{}{}
	for key := range scopeMapping {
		result[key] = false
	}

	for _, scope := range scopes {
		for key, mappedScope := range scopeMapping {
			if scope == mappedScope {
				result[key] = true
				break
			}
		}
	}

	return []interface{}{result}
}
