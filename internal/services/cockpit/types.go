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

func expandCockpitTokenScopes(raw any) []cockpit.TokenScope {
	var expandedScopes []cockpit.TokenScope

	scopesList, ok := raw.([]any)
	if !ok || len(scopesList) == 0 {
		return expandedScopes
	}

	scopesMap, ok := scopesList[0].(map[string]any)
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

func flattenCockpitTokenScopes(scopes []cockpit.TokenScope) []any {
	result := map[string]any{}
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

	return []any{result}
}
