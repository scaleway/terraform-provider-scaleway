package edgeservices

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	edge_services "github.com/scaleway/scaleway-sdk-go/api/edge_services/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func expandS3BackendConfig(raw interface{}) *edge_services.ScalewayS3BackendConfig {
	if raw == nil || len(raw.([]interface{})) != 1 {
		return nil
	}

	rawMap := raw.([]interface{})[0].(map[string]interface{})

	return &edge_services.ScalewayS3BackendConfig{
		BucketName:   types.ExpandStringPtr(rawMap["bucket_name"].(string)),
		BucketRegion: types.ExpandStringPtr(rawMap["bucket_region"].(string)),
		IsWebsite:    types.ExpandBoolPtr(rawMap["is_website"]),
	}
}

func flattenS3BackendConfig(s3backend *edge_services.ScalewayS3BackendConfig) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"bucket_name":   types.FlattenStringPtr(s3backend.BucketName),
			"bucket_region": types.FlattenStringPtr(s3backend.BucketRegion),
			"is_website":    types.FlattenBoolPtr(s3backend.IsWebsite),
		},
	}
}

func expandPurge(raw interface{}) []*edge_services.PurgeRequest {
	if raw == nil {
		return nil
	}

	purgeRequests := []*edge_services.PurgeRequest(nil)

	for _, pr := range raw.(*schema.Set).List() {
		rawPr := pr.(map[string]interface{})
		purgeRequest := &edge_services.PurgeRequest{}
		purgeRequest.PipelineID = rawPr["pipeline_id"].(string)
		purgeRequest.Assets = types.ExpandStringsPtr(rawPr["assets"])
		purgeRequest.All = types.ExpandBoolPtr(rawPr["all"])

		purgeRequests = append(purgeRequests, purgeRequest)
	}

	return purgeRequests
}

func expandTLSSecrets(raw interface{}, region scw.Region) []*edge_services.TLSSecret {
	secrets := []*edge_services.TLSSecret(nil)
	rawSecrets := raw.([]interface{})

	for _, rawSecret := range rawSecrets {
		mapSecret := rawSecret.(map[string]interface{})
		secret := &edge_services.TLSSecret{
			SecretID: locality.ExpandID(mapSecret["secret_id"]),
			Region:   region,
		}
		secrets = append(secrets, secret)
	}

	return secrets
}

func flattenTLSSecrets(secrets []*edge_services.TLSSecret) interface{} {
	if len(secrets) == 0 || secrets == nil {
		return nil
	}

	secretsI := []map[string]interface{}(nil)

	for _, secret := range secrets {
		secretMap := map[string]interface{}{
			"secret_id": secret.SecretID,
			"region":    secret.Region.String(),
		}
		secretsI = append(secretsI, secretMap)
	}

	return secretsI
}

func expandLBBackendConfig(raw interface{}) *edge_services.ScalewayLBBackendConfig {
	lbConfigs := []*edge_services.ScalewayLB(nil)
	rawLbConfigs := raw.([]interface{})

	for _, rawLbConfig := range rawLbConfigs {
		outerMap := rawLbConfig.(map[string]interface{})

		lbConfigList, ok := outerMap["lb_config"].([]interface{})
		if !ok || len(lbConfigList) == 0 {
			continue
		}

		innerMap := lbConfigList[0].(map[string]interface{})
		lbConfig := &edge_services.ScalewayLB{
			ID:         locality.ExpandID(innerMap["id"]),
			Zone:       scw.Zone(innerMap["zone"].(string)),
			FrontendID: locality.ExpandID(innerMap["frontend_id"]),
			IsSsl:      types.ExpandBoolPtr(innerMap["is_ssl"]),
			DomainName: types.ExpandStringPtr(innerMap["domain_name"]),
		}
		lbConfigs = append(lbConfigs, lbConfig)
	}

	return &edge_services.ScalewayLBBackendConfig{
		LBs: lbConfigs,
	}
}

func flattenLBBackendConfig(lbConfigs *edge_services.ScalewayLBBackendConfig) interface{} {
	if lbConfigs == nil {
		return nil
	}

	lbConfigsI := []map[string]interface{}(nil)

	for _, lbConfig := range lbConfigs.LBs {
		secretMap := map[string]interface{}{
			"id":          lbConfig.ID,
			"frontend_id": lbConfig.FrontendID,
			"is_ssl":      types.FlattenBoolPtr(lbConfig.IsSsl),
			"domain_name": types.FlattenStringPtr(lbConfig.DomainName),
			"zone":        lbConfig.Zone.String(),
		}
		lbConfigsI = append(lbConfigsI, secretMap)
	}

	return lbConfigsI
}

func wrapSecretsInConfig(secrets []*edge_services.TLSSecret) *edge_services.TLSSecretsConfig {
	return &edge_services.TLSSecretsConfig{
		TLSSecrets: secrets,
	}
}

func expandRouteRules(raw interface{}) []*edge_services.SetRouteRulesRequestRouteRule {
	if raw == nil {
		return nil
	}

	rulesList := raw.([]interface{})
	result := make([]*edge_services.SetRouteRulesRequestRouteRule, 0, len(rulesList))

	for _, rawRule := range rulesList {
		ruleMap := rawRule.(map[string]interface{})
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

func expandRuleHTTPMatch(raw interface{}) *edge_services.RuleHTTPMatch {
	list, ok := raw.([]interface{})
	if !ok || len(list) < 1 {
		return nil
	}

	ruleMap := list[0].(map[string]interface{})
	result := &edge_services.RuleHTTPMatch{}

	if v, exists := ruleMap["method_filters"]; exists && v != nil {
		filters := v.([]interface{})
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

func expandRuleHTTPMatchPathFilter(raw interface{}) *edge_services.RuleHTTPMatchPathFilter {
	list, ok := raw.([]interface{})
	if !ok || len(list) < 1 {
		return nil
	}

	mapPF := list[0].(map[string]interface{})

	return &edge_services.RuleHTTPMatchPathFilter{
		PathFilterType: edge_services.RuleHTTPMatchPathFilterPathFilterType(mapPF["path_filter_type"].(string)),
		Value:          mapPF["value"].(string),
	}
}

func flattenRouteRules(rules []*edge_services.RouteRule) []interface{} {
	if rules == nil {
		return nil
	}

	result := make([]interface{}, 0, len(rules))

	for _, rule := range rules {
		m := map[string]interface{}{
			"backend_stage_id": types.FlattenStringPtr(rule.BackendStageID),
			"rule_http_match":  flattenRuleHTTPMatch(rule.RuleHTTPMatch),
		}
		result = append(result, m)
	}

	return result
}

func flattenRuleHTTPMatch(match *edge_services.RuleHTTPMatch) []interface{} {
	if match == nil {
		return nil
	}

	m := map[string]interface{}{}

	if match.MethodFilters != nil && len(match.MethodFilters) > 0 {
		filters := make([]interface{}, len(match.MethodFilters))
		for i, v := range match.MethodFilters {
			filters[i] = string(v)
		}
		m["method_filters"] = filters
	} else {
		m["method_filters"] = []interface{}{}
	}

	m["path_filter"] = flattenRuleHTTPMatchPathFilter(match.PathFilter)

	return []interface{}{m}
}

func flattenRuleHTTPMatchPathFilter(pathFilter *edge_services.RuleHTTPMatchPathFilter) []interface{} {
	if pathFilter == nil {
		return nil
	}

	m := map[string]interface{}{
		"path_filter_type": pathFilter.PathFilterType.String(),
		"value":            pathFilter.Value,
	}

	return []interface{}{m}
}
