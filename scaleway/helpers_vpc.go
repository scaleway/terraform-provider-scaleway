package scaleway

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	v1 "github.com/scaleway/scaleway-sdk-go/api/vpc/v1"
	v2 "github.com/scaleway/scaleway-sdk-go/api/vpc/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// vpcAPIWithZone returns a new VPC API and the zone for a Create request
func vpcAPIWithZone(d *schema.ResourceData, m interface{}) (*v1.API, scw.Zone, error) {
	meta := m.(*Meta)
	vpcAPI := v1.NewAPI(meta.scwClient)

	zone, err := extractZone(d, meta)
	if err != nil {
		return nil, "", err
	}
	return vpcAPI, zone, err
}

// vpcAPIWithZoneAndID
func vpcAPIWithZoneAndID(m interface{}, id string) (*v1.API, scw.Zone, string, error) {
	meta := m.(*Meta)
	vpcAPI := v1.NewAPI(meta.scwClient)

	zone, ID, err := parseZonedID(id)
	if err != nil {
		return nil, "", "", err
	}
	return vpcAPI, zone, ID, err
}

// vpcAPIWithRegion returns a new VPC API and the region for a Create request
func vpcAPIWithRegion(d *schema.ResourceData, m interface{}) (*v2.API, scw.Region, error) {
	meta := m.(*Meta)
	vpcAPI := v2.NewAPI(meta.scwClient)

	region, err := extractRegion(d, meta)
	if err != nil {
		return nil, "", err
	}
	return vpcAPI, region, err
}

// vpcAPIWithRegionAndID returns a new VPC API with locality and ID extracted from the state
func vpcAPIWithRegionAndID(m interface{}, id string) (*v2.API, scw.Region, string, error) {
	meta := m.(*Meta)
	vpcAPI := v2.NewAPI(meta.scwClient)

	region, ID, err := parseRegionalID(id)
	if err != nil {
		return nil, "", "", err
	}
	return vpcAPI, region, ID, err
}

func vpcAPI(m interface{}) (*v1.API, error) {
	meta, ok := m.(*Meta)
	if !ok {
		return nil, fmt.Errorf("wrong type: %T", m)
	}

	return v1.NewAPI(meta.scwClient), nil
}

func expandSubnets(d *schema.ResourceData) (ipv4Subnets []scw.IPNet, ipv6Subnets []scw.IPNet, err error) {
	if v, ok := d.GetOk("ipv4_subnet"); ok {
		for _, s := range v.([]interface{}) {
			rawSubnet := s.(map[string]interface{})
			ipNet, err := expandIPNet(rawSubnet["subnet"].(string))
			if err != nil {
				return nil, nil, err
			}
			ipv4Subnets = append(ipv4Subnets, ipNet)
		}
	}

	if v, ok := d.GetOk("ipv6_subnets"); ok {
		for _, s := range v.(*schema.Set).List() {
			rawSubnet := s.(map[string]interface{})
			ipNet, err := expandIPNet(rawSubnet["subnet"].(string))
			if err != nil {
				return nil, nil, err
			}
			ipv6Subnets = append(ipv6Subnets, ipNet)
		}
	}
	return
}

func flattenAndSortSubnets(sub interface{}) (interface{}, interface{}) {
	switch subnets := sub.(type) {
	case []scw.IPNet:
		return flattenAndSortIPNetSubnets(subnets)
	case []*v2.Subnet:
		return flattenAndSortSubnetV2s(subnets)
	default:
		return "", nil
	}
}

func flattenAndSortIPNetSubnets(subnets []scw.IPNet) (interface{}, interface{}) {
	if subnets == nil {
		return "", nil
	}

	flattenedipv4Subnets := []map[string]interface{}(nil)
	flattenedipv6Subnets := []map[string]interface{}(nil)

	for _, s := range subnets {
		// If it's an IPv4 subnet
		if s.IP.To4() != nil {
			sub, err := flattenIPNet(s)
			if err != nil {
				return "", nil
			}
			flattenedipv4Subnets = append(flattenedipv4Subnets, map[string]interface{}{
				"subnet":        sub,
				"address":       s.IP.String(),
				"subnet_mask":   maskHexToDottedDecimal(s.Mask),
				"prefix_length": getPrefixLength(s.Mask),
			})
		} else {
			sub, err := flattenIPNet(s)
			if err != nil {
				return "", nil
			}
			flattenedipv6Subnets = append(flattenedipv6Subnets, map[string]interface{}{
				"subnet":        sub,
				"address":       s.IP.String(),
				"subnet_mask":   maskHexToDottedDecimal(s.IPNet.Mask),
				"prefix_length": getPrefixLength(s.Mask),
			})
		}
	}

	return flattenedipv4Subnets, flattenedipv6Subnets
}

func flattenAndSortSubnetV2s(subnets []*v2.Subnet) (interface{}, interface{}) {
	if subnets == nil {
		return "", nil
	}

	flattenedipv4Subnets := []map[string]interface{}(nil)
	flattenedipv6Subnets := []map[string]interface{}(nil)

	for _, s := range subnets {
		// If it's an IPv4 subnet
		if s.Subnet.IP.To4() != nil {
			sub, err := flattenIPNet(s.Subnet)
			if err != nil {
				return "", nil
			}
			flattenedipv4Subnets = append(flattenedipv4Subnets, map[string]interface{}{
				"id":            s.ID,
				"created_at":    flattenTime(s.CreatedAt),
				"updated_at":    flattenTime(s.UpdatedAt),
				"subnet":        sub,
				"address":       s.Subnet.IP.String(),
				"subnet_mask":   maskHexToDottedDecimal(s.Subnet.Mask),
				"prefix_length": getPrefixLength(s.Subnet.Mask),
			})
		} else {
			sub, err := flattenIPNet(s.Subnet)
			if err != nil {
				return "", nil
			}
			flattenedipv6Subnets = append(flattenedipv6Subnets, map[string]interface{}{
				"id":            s.ID,
				"created_at":    flattenTime(s.CreatedAt),
				"updated_at":    flattenTime(s.UpdatedAt),
				"subnet":        sub,
				"address":       s.Subnet.IP.String(),
				"subnet_mask":   maskHexToDottedDecimal(s.Subnet.Mask),
				"prefix_length": getPrefixLength(s.Subnet.Mask),
			})
		}
	}

	return flattenedipv4Subnets, flattenedipv6Subnets
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
