package domain

import (
	"net"
	"strings"

	domain "github.com/scaleway/scaleway-sdk-go/api/domain/v2beta1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

// FlattenDomainData normalizes domain record data based on record type
func FlattenDomainData(data string, recordType domain.RecordType, dnsZone string) any {
	switch recordType {
	case domain.RecordTypeMX: // API return this format: "{priority} {data}"
		dataSplit := strings.SplitN(data, " ", 2)
		if len(dataSplit) == 2 {
			return dataSplit[1]
		}
	case domain.RecordTypeTXT:
		return strings.Trim(data, "\"")
	case domain.RecordTypeSRV:
		return NormalizeSRVData(data, dnsZone)
	}

	return data
}

// NormalizeSRVData normalizes SRV record data by handling weight field and zone domain suffixes
func NormalizeSRVData(data, dnsZone string) string {
	parts := strings.Fields(data)

	if len(parts) >= 4 {
		priority, weight, port, target := parts[0], parts[1], parts[2], parts[3]
		target = RemoveZoneDomainSuffix(target, dnsZone)

		return strings.Join([]string{priority, weight, port, target}, " ")
	}

	if len(parts) == 3 {
		priority, port, target := parts[0], parts[1], parts[2]
		target = RemoveZoneDomainSuffix(target, dnsZone)

		return strings.Join([]string{priority, "0", port, target}, " ")
	}

	return data
}

// RemoveZoneDomainSuffix removes the zone domain suffix from a target
func RemoveZoneDomainSuffix(target, dnsZone string) string {
	targetNoDot, isAbsolute := strings.CutSuffix(target, ".")
	if !isAbsolute {
		return target
	}

	// Check if target ends with ".dnsZone"
	suffix := "." + dnsZone
	if stripped, ok := strings.CutSuffix(targetNoDot, suffix); ok {
		return stripped
	}

	return target
}

func flattenDomainGeoIP(config *domain.RecordGeoIPConfig) any {
	flattenedResult := []map[string]any{}

	if config == nil {
		return flattenedResult
	}

	flattenedResult = []map[string]any{{}}

	if len(config.Matches) > 0 {
		matches := []map[string]any{}

		for _, match := range config.Matches {
			rawMatch := map[string]any{
				"data": match.Data,
			}
			if len(match.Continents) > 0 {
				rawMatch["continents"] = match.Continents
			}

			if len(match.Countries) > 0 {
				rawMatch["countries"] = match.Countries
			}

			matches = append(matches, rawMatch)
		}

		flattenedResult[0]["matches"] = matches
	}

	return flattenedResult
}

func expandDomainGeoIPConfig(defaultData string, i any, ok bool) *domain.RecordGeoIPConfig {
	if i == nil || !ok {
		return nil
	}

	rawMap := i.([]any)[0].(map[string]any)

	config := domain.RecordGeoIPConfig{
		Default: defaultData,
	}

	rawMatches, ok := rawMap["matches"].([]any)
	if !ok && len(rawMatches) > 0 {
		return &config
	}

	matches := []*domain.RecordGeoIPConfigMatch{}

	for _, rawMatch := range rawMatches {
		rawMatchMap := rawMatch.(map[string]any)

		match := &domain.RecordGeoIPConfigMatch{
			Data: rawMatchMap["data"].(string),
		}

		rawContinents, ok := rawMatchMap["continents"].([]any)
		if ok {
			match.Continents = []string{}
			for _, rawContinent := range rawContinents {
				match.Continents = append(match.Continents, rawContinent.(string))
			}
		}

		rawCountries, ok := rawMatchMap["countries"].([]any)
		if ok {
			match.Countries = []string{}
			for _, rawCountry := range rawCountries {
				match.Countries = append(match.Countries, rawCountry.(string))
			}
		}

		matches = append(matches, match)
	}

	config.Matches = matches

	return &config
}

func flattenDomainHTTPService(config *domain.RecordHTTPServiceConfig) any {
	flattened := []map[string]any{}

	if config == nil {
		return flattened
	}

	ips := []any{}

	if len(config.IPs) > 0 {
		for _, ip := range config.IPs {
			ips = append(ips, ip.String())
		}
	}

	return []map[string]any{
		{
			"must_contain": types.FlattenStringPtr(config.MustContain),
			"url":          config.URL,
			"user_agent":   types.FlattenStringPtr(config.UserAgent),
			"strategy":     config.Strategy.String(),
			"ips":          ips,
		},
	}
}

func expandDomainHTTPService(i any, ok bool) *domain.RecordHTTPServiceConfig {
	if i == nil || !ok {
		return nil
	}

	rawMap := i.([]any)[0].(map[string]any)

	ips := []net.IP{}

	rawIPs, ok := rawMap["ips"].([]any)
	if ok {
		for _, rawIP := range rawIPs {
			ips = append(ips, net.ParseIP(rawIP.(string)))
		}
	}

	return &domain.RecordHTTPServiceConfig{
		MustContain: types.ExpandStringPtr(rawMap["must_contain"]),
		URL:         rawMap["url"].(string),
		UserAgent:   types.ExpandStringPtr(rawMap["user_agent"]),
		Strategy:    domain.RecordHTTPServiceConfigStrategy(rawMap["strategy"].(string)),
		IPs:         ips,
	}
}

func flattenDomainWeighted(config *domain.RecordWeightedConfig) any {
	flattened := []map[string]any{}

	if config == nil {
		return flattened
	}

	if len(config.WeightedIPs) > 0 {
		for _, weightedIPs := range config.WeightedIPs {
			flattened = append(flattened, map[string]any{
				"ip":     weightedIPs.IP.String(),
				"weight": int(weightedIPs.Weight),
			})
		}
	}

	return flattened
}

func expandDomainWeighted(i any, ok bool) *domain.RecordWeightedConfig {
	if i == nil || !ok {
		return nil
	}

	weightedIPs := []*domain.RecordWeightedConfigWeightedIP{}

	if raw := i.([]any); len(raw) > 0 {
		for _, rawWeighted := range raw {
			rawMap := rawWeighted.(map[string]any)
			weightedIPs = append(weightedIPs, &domain.RecordWeightedConfigWeightedIP{
				IP:     net.ParseIP(rawMap["ip"].(string)),
				Weight: uint32(rawMap["weight"].(int)),
			})
		}
	}

	return &domain.RecordWeightedConfig{
		WeightedIPs: weightedIPs,
	}
}

func flattenDomainView(config *domain.RecordViewConfig) any {
	flattened := []map[string]any{}

	if config == nil {
		return flattened
	}

	if len(config.Views) > 0 {
		for _, view := range config.Views {
			flattened = append(flattened, map[string]any{
				"subnet": view.Subnet,
				"data":   view.Data,
			})
		}
	}

	return flattened
}

func expandDomainView(i any, ok bool) *domain.RecordViewConfig {
	if i == nil || !ok {
		return nil
	}

	views := []*domain.RecordViewConfigView{}

	if raw := i.([]any); len(raw) > 0 {
		for _, rawWeighted := range raw {
			rawMap := rawWeighted.(map[string]any)
			views = append(views, &domain.RecordViewConfigView{
				Subnet: rawMap["subnet"].(string),
				Data:   rawMap["data"].(string),
			})
		}
	}

	return &domain.RecordViewConfig{
		Views: views,
	}
}
