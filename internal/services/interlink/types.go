package interlink

import (
	interlink "github.com/scaleway/scaleway-sdk-go/api/interlink/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

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