package flexibleip

import (
	flexibleip "github.com/scaleway/scaleway-sdk-go/api/flexibleip/v1alpha1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func flattenFlexibleIPMacAddress(mac *flexibleip.MACAddress) any {
	if mac == nil {
		return nil
	}

	return []map[string]any{
		{
			"id":          mac.ID,
			"mac_address": mac.MacAddress,
			"mac_type":    mac.MacType,
			"status":      mac.Status,
			"created_at":  types.FlattenTime(mac.CreatedAt),
			"updated_at":  types.FlattenTime(mac.UpdatedAt),
			"zone":        mac.Zone,
		},
	}
}
