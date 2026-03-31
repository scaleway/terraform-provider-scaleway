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
