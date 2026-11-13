package cockpit

import (
	"fmt"
	"slices"

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

// defaultCockpitTokenScopes returns the scopes used when the scopes block is omitted
// (write_metrics and write_logs), matching the managed scaleway_cockpit_token defaults.
func defaultCockpitTokenScopes() []cockpit.TokenScope {
	return []cockpit.TokenScope{
		cockpit.TokenScopeWriteOnlyMetrics,
		cockpit.TokenScopeWriteOnlyLogs,
	}
}

// expandCockpitTokenScopesFromFlags expands a name→enabled map using scopeMapping.
// Shared by the managed and the ephemeral token resources.
func expandCockpitTokenScopesFromFlags(flags map[string]any) []cockpit.TokenScope {
	var expandedScopes []cockpit.TokenScope

	for key, tokenScope := range scopeMapping {
		if value, ok := flags[key].(bool); ok && value {
			expandedScopes = append(expandedScopes, tokenScope)
		}
	}

	return expandedScopes
}

// flattenCockpitTokenScopes returns the enabled state of every scope flag.
func flattenCockpitTokenScopes(scopes []cockpit.TokenScope) map[string]bool {
	result := make(map[string]bool, len(scopeMapping))
	for key, mappedScope := range scopeMapping {
		result[key] = slices.Contains(scopes, mappedScope)
	}

	return result
}
