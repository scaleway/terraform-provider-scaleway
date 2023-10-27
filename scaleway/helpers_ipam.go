package scaleway

import (
	"net"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	ipam "github.com/scaleway/scaleway-sdk-go/api/ipam/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/scaleway-sdk-go/validation"
)

// ipamAPIWithRegion returns a new ipam API and the region
func ipamAPIWithRegion(d *schema.ResourceData, m interface{}) (*ipam.API, scw.Region, error) {
	meta := m.(*Meta)
	ipamAPI := ipam.NewAPI(meta.scwClient)

	region, err := extractRegion(d, meta)
	if err != nil {
		return nil, "", err
	}

	return ipamAPI, region, nil
}

// ipamAPIWithRegionAndID returns a new ipam API with locality and ID extracted from the state
func ipamAPIWithRegionAndID(m interface{}, id string) (*ipam.API, scw.Region, string, error) {
	meta := m.(*Meta)
	ipamAPI := ipam.NewAPI(meta.scwClient)

	region, ID, err := parseRegionalID(id)
	if err != nil {
		return nil, "", "", err
	}
	return ipamAPI, region, ID, err
}

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
		Zonal:            expandStringPtr(rawMap["zonal"].(string)),
		PrivateNetworkID: expandStringPtr(expandID(rawMap["private_network_id"].(string))),
		SubnetID:         expandStringPtr(rawMap["subnet_id"].(string)),
	}
}

func flattenIPSource(source *ipam.Source, privateNetworkID string) interface{} {
	if source == nil {
		return nil
	}
	return []map[string]interface{}{
		{
			"zonal":              flattenStringPtr(source.Zonal),
			"private_network_id": privateNetworkID,
			"subnet_id":          flattenStringPtr(source.SubnetID),
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
			"mac_address": flattenStringPtr(resource.MacAddress),
			"name":        flattenStringPtr(resource.Name),
		},
	}
}

func checkSubnetIDInFlattenedSubnets(subnetID string, flattenedSubnets interface{}) bool {
	for _, subnet := range flattenedSubnets.([]map[string]interface{}) {
		if subnet["id"].(string) == subnetID {
			return true
		}
	}
	return false
}

func diffSuppressFuncStandaloneIPandCIDR(k, old, new string, _ *schema.ResourceData) bool {
	oldIP, oldNet, errOld := net.ParseCIDR(old)
	if errOld != nil {
		oldIP = net.ParseIP(old)
	}

	newIP, newNet, errNew := net.ParseCIDR(new)
	if errNew != nil {
		newIP = net.ParseIP(new)
	}

	if oldIP != nil && newIP != nil && oldIP.Equal(newIP) {
		return true
	}

	if oldNet != nil && newIP != nil && oldNet.Contains(newIP) {
		return true
	}

	if newNet != nil && oldIP != nil && newNet.Contains(oldIP) {
		return true
	}

	return false
}
