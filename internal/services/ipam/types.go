//nolint:dupword
package ipam

import (
	"strings"

	"github.com/scaleway/scaleway-sdk-go/api/ipam/v1"
	"github.com/scaleway/scaleway-sdk-go/validation"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

// expandLastID expand the last ID in a potential composed ID
// region/id1/id2 -> id2
// region/id1 -> id1
// region/id1/invalid -> id1
// id1 -> id1
// invalid -> invalid
func expandLastID(i any) string {
	composedID := i.(string)

	elems := strings.Split(composedID, "/")
	for i := len(elems) - 1; i >= 0; i-- {
		if validation.IsUUID(elems[i]) {
			return elems[i]
		}
	}

	return composedID
}

func expandIPSource(raw any) *ipam.Source {
	if raw == nil || len(raw.([]any)) != 1 {
		return nil
	}

	rawMap := raw.([]any)[0].(map[string]any)

	return &ipam.Source{
		Zonal:            types.ExpandStringPtr(rawMap["zonal"].(string)),
		PrivateNetworkID: types.ExpandStringPtr(locality.ExpandID(rawMap["private_network_id"].(string))),
		SubnetID:         types.ExpandStringPtr(rawMap["subnet_id"].(string)),
	}
}

func expandCustomResource(raw any) *ipam.CustomResource {
	if raw == nil || len(raw.([]any)) != 1 {
		return nil
	}

	rawMap := raw.([]any)[0].(map[string]any)

	return &ipam.CustomResource{
		MacAddress: rawMap["mac_address"].(string),
		Name:       types.ExpandStringPtr(rawMap["name"].(string)),
	}
}

func flattenIPSource(source *ipam.Source, privateNetworkID string) any {
	if source == nil {
		return nil
	}

	return []map[string]any{
		{
			"zonal":              types.FlattenStringPtr(source.Zonal),
			"private_network_id": privateNetworkID,
			"subnet_id":          types.FlattenStringPtr(source.SubnetID),
		},
	}
}

func flattenIPResource(resource *ipam.Resource) any {
	if resource == nil {
		return nil
	}

	return []map[string]any{
		{
			"type":        resource.Type.String(),
			"id":          resource.ID,
			"mac_address": types.FlattenStringPtr(resource.MacAddress),
			"name":        types.FlattenStringPtr(resource.Name),
		},
	}
}

func flattenIPReverse(reverse *ipam.Reverse) any {
	if reverse == nil {
		return nil
	}

	return map[string]any{
		"hostname": reverse.Hostname,
		"address":  types.FlattenIPPtr(reverse.Address),
	}
}

func flattenIPReverses(reverses []*ipam.Reverse) any {
	rawReverses := make([]any, 0, len(reverses))
	for _, reverse := range reverses {
		rawReverses = append(rawReverses, flattenIPReverse(reverse))
	}

	return rawReverses
}

func checkSubnetIDInFlattenedSubnets(subnetID string, flattenedSubnets any) bool {
	for _, subnet := range flattenedSubnets.([]map[string]any) {
		if subnet["id"].(string) == subnetID {
			return true
		}
	}

	return false
}
