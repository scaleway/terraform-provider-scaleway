package edgeservices

import (
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	edge_services "github.com/scaleway/scaleway-sdk-go/api/edge_services/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func expandS3BackendConfig(raw any) *edge_services.ScalewayS3BackendConfig {
	if raw == nil || len(raw.([]any)) != 1 {
		return nil
	}

	rawMap := raw.([]any)[0].(map[string]any)

	return &edge_services.ScalewayS3BackendConfig{
		BucketName:   types.ExpandStringPtr(rawMap["bucket_name"].(string)),
		BucketRegion: types.ExpandStringPtr(rawMap["bucket_region"].(string)),
		IsWebsite:    types.ExpandBoolPtr(rawMap["is_website"]),
	}
}

func flattenS3BackendConfig(s3backend *edge_services.ScalewayS3BackendConfig) []map[string]any {
	return []map[string]any{
		{
			"bucket_name":   types.FlattenStringPtr(s3backend.BucketName),
			"bucket_region": types.FlattenStringPtr(s3backend.BucketRegion),
			"is_website":    types.FlattenBoolPtr(s3backend.IsWebsite),
		},
	}
}

func expandPurge(raw any) []*edge_services.PurgeRequest {
	if raw == nil {
		return nil
	}

	purgeRequests := []*edge_services.PurgeRequest(nil)

	for _, pr := range raw.(*schema.Set).List() {
		rawPr := pr.(map[string]any)
		purgeRequest := &edge_services.PurgeRequest{}
		purgeRequest.PipelineID = rawPr["pipeline_id"].(string)
		purgeRequest.Assets = types.ExpandStringsPtr(rawPr["assets"])
		purgeRequest.All = types.ExpandBoolPtr(rawPr["all"])

		purgeRequests = append(purgeRequests, purgeRequest)
	}

	return purgeRequests
}

func expandTLSSecrets(raw any, region scw.Region) []*edge_services.TLSSecret {
	rawSecrets := raw.([]any)
	secrets := make([]*edge_services.TLSSecret, 0, len(rawSecrets))

	for _, rawSecret := range rawSecrets {
		mapSecret := rawSecret.(map[string]any)
		secret := &edge_services.TLSSecret{
			SecretID: locality.ExpandID(mapSecret["secret_id"]),
			Region:   region,
		}
		secrets = append(secrets, secret)
	}

	return secrets
}

func flattenTLSSecrets(secrets []*edge_services.TLSSecret) any {
	if len(secrets) == 0 || secrets == nil {
		return nil
	}

	secretsI := []map[string]any(nil)

	for _, secret := range secrets {
		secretMap := map[string]any{
			"secret_id": secret.SecretID,
			"region":    secret.Region.String(),
		}
		secretsI = append(secretsI, secretMap)
	}

	return secretsI
}

func expandLBBackendConfig(d *schema.ResourceData, zone scw.Zone, raw any) *edge_services.ScalewayLBBackendConfig {
	lbConfigs := []*edge_services.ScalewayLB(nil)
	rawLbConfigs := raw.([]any)

	for _, rawLbConfig := range rawLbConfigs {
		outerMap := rawLbConfig.(map[string]any)

		lbConfigList, ok := outerMap["lb_config"].([]any)
		if !ok || len(lbConfigList) == 0 {
			continue
		}

		innerMap := lbConfigList[0].(map[string]any)

		configZone := zone
		if rawZone, ok := meta.GetRawConfigForKey(d, "lb_backend_config.0.lb_config.0.zone", cty.String); ok && rawZone != nil && rawZone != "" {
			configZone = scw.Zone(rawZone.(string))
		} else if lbID, ok := innerMap["id"].(string); ok && lbID != "" {
			if zonalID := zonal.ExpandID(lbID); zonalID.Zone != "" {
				configZone = zonalID.Zone
			}
		}

		lbConfig := &edge_services.ScalewayLB{
			ID:           locality.ExpandID(innerMap["id"]),
			Zone:         configZone,
			FrontendID:   locality.ExpandID(innerMap["frontend_id"]),
			IsSsl:        types.ExpandBoolPtr(innerMap["is_ssl"]),
			DomainName:   types.ExpandStringPtr(innerMap["domain_name"]),
			HasWebsocket: types.ExpandBoolPtr(innerMap["has_websocket"]),
		}
		lbConfigs = append(lbConfigs, lbConfig)
	}

	return &edge_services.ScalewayLBBackendConfig{
		LBs: lbConfigs,
	}
}

func flattenLBBackendConfig(zone scw.Zone, lbConfigs *edge_services.ScalewayLBBackendConfig) any {
	if lbConfigs == nil {
		return nil
	}

	inner := make([]any, len(lbConfigs.LBs))

	for i, lbConfig := range lbConfigs.LBs {
		configZone := lbConfig.Zone
		if configZone == "" {
			configZone = zone
		}

		inner[i] = map[string]any{
			"id":            zonal.NewIDString(configZone, lbConfig.ID),
			"frontend_id":   zonal.NewIDString(configZone, lbConfig.FrontendID),
			"is_ssl":        types.FlattenBoolPtr(lbConfig.IsSsl),
			"domain_name":   types.FlattenStringPtr(lbConfig.DomainName),
			"zone":          configZone.String(),
			"has_websocket": types.FlattenBoolPtr(lbConfig.HasWebsocket),
		}
	}

	outer := []map[string]any{{
		"lb_config": inner,
	}}

	return outer
}

func wrapSecretsInConfig(secrets []*edge_services.TLSSecret) *edge_services.TLSSecretsConfig {
	return &edge_services.TLSSecretsConfig{
		TLSSecrets: secrets,
	}
}

func expandRouteRules(raw any) []*edge_services.SetRouteRulesRequestRouteRule {
	if raw == nil {
		return nil
	}

	rulesList := raw.([]any)
	result := make([]*edge_services.SetRouteRulesRequestRouteRule, 0, len(rulesList))

	for _, rawRule := range rulesList {
		ruleMap := rawRule.(map[string]any)
		rule := &edge_services.SetRouteRulesRequestRouteRule{
			BackendStageID: types.ExpandStringPtr(ruleMap["backend_stage_id"].(string)),
		}

		if rawHTTPMatch, ok := ruleMap["rule_http_match"]; ok && rawHTTPMatch != nil {
			if expandedHTTP := expandRuleHTTPMatch(rawHTTPMatch); expandedHTTP != nil {
				rule.RuleHTTPMatch = expandedHTTP
			}
		}

		result = append(result, rule)
	}

	return result
}

func expandRuleHTTPMatch(raw any) *edge_services.RuleHTTPMatch {
	list, ok := raw.([]any)
	if !ok || len(list) < 1 {
		return nil
	}

	ruleMap := list[0].(map[string]any)
	result := &edge_services.RuleHTTPMatch{}

	if v, exists := ruleMap["method_filters"]; exists && v != nil {
		filters := v.([]any)
		result.MethodFilters = make([]edge_services.RuleHTTPMatchMethodFilter, len(filters))

		for i, item := range filters {
			result.MethodFilters[i] = edge_services.RuleHTTPMatchMethodFilter(item.(string))
		}
	}

	if rawPF, exists := ruleMap["path_filter"]; exists && rawPF != nil {
		result.PathFilter = expandRuleHTTPMatchPathFilter(rawPF)
	}

	return result
}

func expandRuleHTTPMatchPathFilter(raw any) *edge_services.RuleHTTPMatchPathFilter {
	list, ok := raw.([]any)
	if !ok || len(list) < 1 {
		return nil
	}

	mapPF := list[0].(map[string]any)

	return &edge_services.RuleHTTPMatchPathFilter{
		PathFilterType: edge_services.RuleHTTPMatchPathFilterPathFilterType(mapPF["path_filter_type"].(string)),
		Value:          mapPF["value"].(string),
	}
}

func flattenRouteRules(rules []*edge_services.RouteRule) []any {
	if rules == nil {
		return nil
	}

	result := make([]any, 0, len(rules))

	for _, rule := range rules {
		m := map[string]any{
			"backend_stage_id": types.FlattenStringPtr(rule.BackendStageID),
			"rule_http_match":  flattenRuleHTTPMatch(rule.RuleHTTPMatch),
		}
		result = append(result, m)
	}

	return result
}

func flattenRuleHTTPMatch(match *edge_services.RuleHTTPMatch) []any {
	if match == nil {
		return nil
	}

	m := map[string]any{}

	if len(match.MethodFilters) > 0 {
		filters := make([]any, len(match.MethodFilters))
		for i, v := range match.MethodFilters {
			filters[i] = string(v)
		}

		m["method_filters"] = filters
	} else {
		m["method_filters"] = []any{}
	}

	m["path_filter"] = flattenRuleHTTPMatchPathFilter(match.PathFilter)

	return []any{m}
}

func flattenRuleHTTPMatchPathFilter(pathFilter *edge_services.RuleHTTPMatchPathFilter) []any {
	if pathFilter == nil {
		return nil
	}

	m := map[string]any{
		"path_filter_type": pathFilter.PathFilterType.String(),
		"value":            pathFilter.Value,
	}

	return []any{m}
}
