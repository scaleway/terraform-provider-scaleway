package cockpit

import (
	"fmt"

	"github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
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

func flattenCockpitEndpoints(dataSources []*cockpit.DataSource, grafanaURL string, alertManagerURL string) []map[string]interface{} {
	endpointMap := map[string]string{}

	for _, dataSource := range dataSources {
		switch dataSource.Type {
		case "metrics":
			endpointMap["metrics_url"] = dataSource.URL
		case "logs":
			endpointMap["logs_url"] = dataSource.URL
		case "traces":
			endpointMap["traces_url"] = dataSource.URL
		}
	}

	endpoints := []map[string]interface{}{
		{
			"metrics_url":      endpointMap["metrics_url"],
			"logs_url":         endpointMap["logs_url"],
			"alertmanager_url": alertManagerURL,
			"grafana_url":      grafanaURL,
			"traces_url":       endpointMap["traces_url"],
		},
	}

	return endpoints
}

func createCockpitPushURLList(endpoints []map[string]interface{}) []map[string]interface{} {
	var result []map[string]interface{}

	for _, endpoint := range endpoints {
		newEndpoint := make(map[string]interface{})

		if metricsURL, ok := endpoint["metrics_url"].(string); ok && metricsURL != "" {
			newEndpoint["push_metrics_url"] = metricsURL + pathMetricsURL
		}

		if logsURL, ok := endpoint["logs_url"].(string); ok && logsURL != "" {
			newEndpoint["push_logs_url"] = logsURL + pathLogsURL
		}

		if len(newEndpoint) > 0 {
			result = append(result, newEndpoint)
		}
	}
	return result
}

func createCockpitPushURL(sourceType cockpit.DataSourceType, url string) (string, error) {
	switch sourceType {
	case cockpit.DataSourceTypeMetrics:
		return url + pathMetricsURL, nil
	case cockpit.DataSourceTypeLogs:
		return url + pathLogsURL, nil
	case cockpit.DataSourceTypeTraces:
		return url + pathTracesURL, nil
	default:
		return "", fmt.Errorf("invalid data source type: %v", sourceType)
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
