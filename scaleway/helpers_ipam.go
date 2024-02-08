//nolint:dupword
package scaleway

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	ipam "github.com/scaleway/scaleway-sdk-go/api/ipam/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/scaleway-sdk-go/validation"
)

const (
	defaultIPAMIPReverseDNSTimeout = 10 * time.Minute
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

func flattenIPReverse(reverse *ipam.Reverse) interface{} {
	if reverse == nil {
		return nil
	}

	return map[string]interface{}{
		"hostname": reverse.Hostname,
		"address":  flattenIPPtr(reverse.Address),
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

func diffSuppressFuncStandaloneIPandCIDR(_, oldValue, newValue string, _ *schema.ResourceData) bool {
	oldIP, oldNet, errOld := net.ParseCIDR(oldValue)
	if errOld != nil {
		oldIP = net.ParseIP(oldValue)
	}

	newIP, newNet, errNew := net.ParseCIDR(newValue)
	if errNew != nil {
		newIP = net.ParseIP(newValue)
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

func validateStandaloneIPorDefaultCIDR() func(interface{}, string) ([]string, []error) {
	return func(val interface{}, key string) (warns []string, errs []error) {
		ip, isString := val.(string)
		if !isString {
			return nil, []error{fmt.Errorf("invalid input for key '%s': not a string", key)}
		}

		// Check if it's a standalone IP address
		if net.ParseIP(ip) != nil {
			return
		}

		// Check if it's an IP with CIDR notation
		_, ipNet, err := net.ParseCIDR(ip)
		if err != nil {
			errs = append(errs, fmt.Errorf("%q is not a valid IP address or CIDR notation: %s", key, ip))
			return
		}
		ones, _ := ipNet.Mask.Size()
		if (ipNet.IP.To4() != nil && ones != 32) || (ipNet.IP.To16() != nil && ipNet.IP.To4() == nil && ones != 128) {
			errs = append(errs, fmt.Errorf("%q must be a /32 CIDR notation for IPv4 or /128 for IPv6: %s", key, ip))
		}

		return
	}
}
