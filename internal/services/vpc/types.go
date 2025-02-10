package vpc

import (
	"net"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/vpc/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func expandSubnets(d *schema.ResourceData) (ipv4Subnets []scw.IPNet, ipv6Subnets []scw.IPNet, err error) {
	if v, ok := d.GetOk("ipv4_subnet"); ok {
		for _, s := range v.([]interface{}) {
			rawSubnet := s.(map[string]interface{})
			ipNet, err := types.ExpandIPNet(rawSubnet["subnet"].(string))
			if err != nil {
				return nil, nil, err
			}
			ipv4Subnets = append(ipv4Subnets, ipNet)
		}
	}

	if v, ok := d.GetOk("ipv6_subnets"); ok {
		for _, s := range v.(*schema.Set).List() {
			rawSubnet := s.(map[string]interface{})
			ipNet, err := types.ExpandIPNet(rawSubnet["subnet"].(string))
			if err != nil {
				return nil, nil, err
			}
			ipv6Subnets = append(ipv6Subnets, ipNet)
		}
	}

	return
}

func FlattenAndSortSubnets(sub interface{}) (interface{}, interface{}) {
	switch subnets := sub.(type) {
	case []scw.IPNet:
		return flattenAndSortIPNetSubnets(subnets)
	case []*vpc.Subnet:
		return flattenAndSortSubnetV2s(subnets)
	default:
		return "", nil
	}
}

func flattenAndSortIPNetSubnets(subnets []scw.IPNet) (interface{}, interface{}) {
	if subnets == nil {
		return "", nil
	}

	flatIpv4Subnets := []map[string]interface{}(nil)
	flatIpv6Subnets := []map[string]interface{}(nil)

	for _, s := range subnets {
		// If it's an IPv4 subnet
		if s.IP.To4() != nil {
			sub, err := types.FlattenIPNet(s)
			if err != nil {
				return "", nil
			}
			flatIpv4Subnets = append(flatIpv4Subnets, map[string]interface{}{
				"subnet":        sub,
				"address":       s.IP.String(),
				"subnet_mask":   maskHexToDottedDecimal(s.Mask),
				"prefix_length": getPrefixLength(s.Mask),
			})
		} else {
			sub, err := types.FlattenIPNet(s)
			if err != nil {
				return "", nil
			}
			flatIpv6Subnets = append(flatIpv6Subnets, map[string]interface{}{
				"subnet":        sub,
				"address":       s.IP.String(),
				"subnet_mask":   maskHexToDottedDecimal(s.IPNet.Mask),
				"prefix_length": getPrefixLength(s.Mask),
			})
		}
	}

	return flatIpv4Subnets, flatIpv6Subnets
}

func flattenAndSortSubnetV2s(subnets []*vpc.Subnet) (interface{}, interface{}) {
	if subnets == nil {
		return "", nil
	}

	flatIpv4Subnets := []map[string]interface{}(nil)
	flatIpv6Subnets := []map[string]interface{}(nil)

	for _, s := range subnets {
		// If it's an IPv4 subnet
		if s.Subnet.IP.To4() != nil {
			sub, err := types.FlattenIPNet(s.Subnet)
			if err != nil {
				return "", nil
			}
			flatIpv4Subnets = append(flatIpv4Subnets, map[string]interface{}{
				"id":            s.ID,
				"created_at":    types.FlattenTime(s.CreatedAt),
				"updated_at":    types.FlattenTime(s.UpdatedAt),
				"subnet":        sub,
				"address":       s.Subnet.IP.String(),
				"subnet_mask":   maskHexToDottedDecimal(s.Subnet.Mask),
				"prefix_length": getPrefixLength(s.Subnet.Mask),
			})
		} else {
			sub, err := types.FlattenIPNet(s.Subnet)
			if err != nil {
				return "", nil
			}
			flatIpv6Subnets = append(flatIpv6Subnets, map[string]interface{}{
				"id":            s.ID,
				"created_at":    types.FlattenTime(s.CreatedAt),
				"updated_at":    types.FlattenTime(s.UpdatedAt),
				"subnet":        sub,
				"address":       s.Subnet.IP.String(),
				"subnet_mask":   maskHexToDottedDecimal(s.Subnet.Mask),
				"prefix_length": getPrefixLength(s.Subnet.Mask),
			})
		}
	}

	return flatIpv4Subnets, flatIpv6Subnets
}

func maskHexToDottedDecimal(mask net.IPMask) string {
	if len(mask) != net.IPv4len && len(mask) != net.IPv6len {
		return ""
	}

	parts := make([]string, len(mask))
	for i, part := range mask {
		parts[i] = strconv.Itoa(int(part))
	}

	return strings.Join(parts, ".")
}

func getPrefixLength(mask net.IPMask) int {
	ones, _ := mask.Size()

	return ones
}
