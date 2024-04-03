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
func expandLastID(i interface{}) string {
	composedID := i.(string)
	elems := strings.Split(composedID, "/")
	for i := len(elems) - 1; i >= 0; i-- {
		if validation.IsUUID(elems[i]) {
			return elems[i]
		}
	}

	return composedID
}

func expandIPSource(raw interface{}) *ipam.Source {
	if raw == nil || len(raw.([]interface{})) != 1 {
		return nil
	}

	rawMap := raw.([]interface{})[0].(map[string]interface{})
	return &ipam.Source{
		Zonal:            types.ExpandStringPtr(rawMap["zonal"].(string)),
		PrivateNetworkID: types.ExpandStringPtr(locality.ExpandID(rawMap["private_network_id"].(string))),
		SubnetID:         types.ExpandStringPtr(rawMap["subnet_id"].(string)),
	}
}

func flattenIPSource(source *ipam.Source, privateNetworkID string) interface{} {
	if source == nil {
		return nil
	}
	return []map[string]interface{}{
		{
			"zonal":              types.FlattenStringPtr(source.Zonal),
			"private_network_id": privateNetworkID,
			"subnet_id":          types.FlattenStringPtr(source.SubnetID),
		},
	}
}

func flattenIPResource(resource *ipam.Resource) interface{} {
	if resource == nil {
		return nil
	}
	return []map[string]interface{}{
		{
			"type":        resource.Type.String(),
			"id":          resource.ID,
			"mac_address": types.FlattenStringPtr(resource.MacAddress),
			"name":        types.FlattenStringPtr(resource.Name),
		},
	}
}

func flattenIPReverse(reverse *ipam.Reverse) interface{} {
	if reverse == nil {
		return nil
	}

	return map[string]interface{}{
		"hostname": reverse.Hostname,
		"address":  types.FlattenIPPtr(reverse.Address),
	}
}

func flattenIPReverses(reverses []*ipam.Reverse) interface{} {
	rawReverses := make([]interface{}, 0, len(reverses))
	for _, reverse := range reverses {
		rawReverses = append(rawReverses, flattenIPReverse(reverse))
	}
	return rawReverses
}

func checkSubnetIDInFlattenedSubnets(subnetID string, flattenedSubnets interface{}) bool {
	for _, subnet := range flattenedSubnets.([]map[string]interface{}) {
		if subnet["id"].(string) == subnetID {
			return true
		}
	}
	return false
}
