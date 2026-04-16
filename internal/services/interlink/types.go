package interlink

import (
	interlink "github.com/scaleway/scaleway-sdk-go/api/interlink/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func expandPrefixFilters(raw any) ([]scw.IPNet, error) {
	if raw == nil {
		return nil, nil
	}

	rawList, ok := raw.([]any)
	if !ok || len(rawList) == 0 {
		return nil, nil
	}

	prefixes := make([]scw.IPNet, 0, len(rawList))
	for _, v := range rawList {
		ipNet, err := types.ExpandIPNet(v.(string))
		if err != nil {
			return nil, err
		}

		prefixes = append(prefixes, ipNet)
	}

	return prefixes, nil
}

func flattenPrefixFilters(prefixes []scw.IPNet) ([]string, error) {
	res := make([]string, 0, len(prefixes))

	for _, p := range prefixes {
		flattened, err := types.FlattenIPNet(p)
		if err != nil {
			return nil, err
		}

		res = append(res, flattened)
	}

	return res, nil
}

func flattenPartners(region scw.Region, partners []*interlink.Partner) []map[string]any {
	result := make([]map[string]any, len(partners))

	for i, partner := range partners {
		result[i] = map[string]any{
			"id":            regional.NewIDString(region, partner.ID),
			"name":          partner.Name,
			"contact_email": partner.ContactEmail,
			"logo_url":      partner.LogoURL,
			"portal_url":    partner.PortalURL,
			"created_at":    types.FlattenTime(partner.CreatedAt),
			"updated_at":    types.FlattenTime(partner.UpdatedAt),
		}
	}

	return result
}

func flattenBgpConfig(config *interlink.BgpConfig) (any, error) {
	if config == nil {
		return nil, nil
	}

	ipv4, err := types.FlattenIPNet(config.IPv4)
	if err != nil {
		return nil, err
	}

	ipv6, err := types.FlattenIPNet(config.IPv6)
	if err != nil {
		return nil, err
	}

	return []map[string]any{
		{
			"asn":  int(config.Asn),
			"ipv4": ipv4,
			"ipv6": ipv6,
		},
	}, nil
}

func flattenPops(pops []*interlink.Pop) []map[string]any {
	result := make([]map[string]any, len(pops))

	for i, pop := range pops {
		bandwidths := make([]int, len(pop.AvailableLinkBandwidthsMbps))
		for j, b := range pop.AvailableLinkBandwidthsMbps {
			bandwidths[j] = int(b)
		}

		result[i] = map[string]any{
			"id":                             regional.NewIDString(pop.Region, pop.ID),
			"name":                           pop.Name,
			"hosting_provider_name":          pop.HostingProviderName,
			"address":                        pop.Address,
			"city":                           pop.City,
			"logo_url":                       pop.LogoURL,
			"available_link_bandwidths_mbps": bandwidths,
			"display_name":                   pop.DisplayName,
			"region":                         pop.Region.String(),
		}
	}

	return result
}
