package vpc

import (
	"net"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/vpc/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func expandSubnets(d *schema.ResourceData) (ipv4Subnets []scw.IPNet, ipv6Subnets []scw.IPNet, err error) {
	if v, ok := d.GetOk("ipv4_subnet"); ok {
		for _, s := range v.([]any) {
			rawSubnet := s.(map[string]any)

			ipNet, err := types.ExpandIPNet(rawSubnet["subnet"].(string))
			if err != nil {
				return nil, nil, err
			}

			ipv4Subnets = append(ipv4Subnets, ipNet)
		}
	}

	if v, ok := d.GetOk("ipv6_subnets"); ok {
		for _, s := range v.(*schema.Set).List() {
			rawSubnet := s.(map[string]any)

			ipNet, err := types.ExpandIPNet(rawSubnet["subnet"].(string))
			if err != nil {
				return nil, nil, err
			}

			ipv6Subnets = append(ipv6Subnets, ipNet)
		}
	}

	return
}

func FlattenAndSortSubnets(sub any) (any, any) {
	switch subnets := sub.(type) {
	case []scw.IPNet:
		return flattenAndSortIPNetSubnets(subnets)
	case []*vpc.Subnet:
		return flattenAndSortSubnetV2s(subnets)
	default:
		return "", nil
	}
}

func flattenAndSortIPNetSubnets(subnets []scw.IPNet) (any, any) {
	if subnets == nil {
		return "", nil
	}

	flatIpv4Subnets := []map[string]any(nil)
	flatIpv6Subnets := []map[string]any(nil)

	for _, s := range subnets {
		// If it's an IPv4 subnet
		if s.IP.To4() != nil {
			sub, err := types.FlattenIPNet(s)
			if err != nil {
				return "", nil
			}

			flatIpv4Subnets = append(flatIpv4Subnets, map[string]any{
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

			flatIpv6Subnets = append(flatIpv6Subnets, map[string]any{
				"subnet":        sub,
				"address":       s.IP.String(),
				"subnet_mask":   maskHexToDottedDecimal(s.Mask),
				"prefix_length": getPrefixLength(s.Mask),
			})
		}
	}

	return flatIpv4Subnets, flatIpv6Subnets
}

func flattenAndSortSubnetV2s(subnets []*vpc.Subnet) (any, any) {
	if subnets == nil {
		return "", nil
	}

	flatIpv4Subnets := []map[string]any(nil)
	flatIpv6Subnets := []map[string]any(nil)

	for _, s := range subnets {
		// If it's an IPv4 subnet
		if s.Subnet.IP.To4() != nil {
			sub, err := types.FlattenIPNet(s.Subnet)
			if err != nil {
				return "", nil
			}

			flatIpv4Subnets = append(flatIpv4Subnets, map[string]any{
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

			flatIpv6Subnets = append(flatIpv6Subnets, map[string]any{
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

func expandACLRules(data any) ([]*vpc.ACLRule, error) {
	if data == nil {
		return nil, nil
	}

	rules := []*vpc.ACLRule(nil)

	for _, rule := range data.([]any) {
		rawRule := rule.(map[string]any)
		ACLRule := &vpc.ACLRule{}

		source, err := types.ExpandIPNet(rawRule["source"].(string))
		if err != nil {
			return nil, err
		}

		destination, err := types.ExpandIPNet(rawRule["destination"].(string))
		if err != nil {
			return nil, err
		}

		ACLRule.Protocol = vpc.ACLRuleProtocol(rawRule["protocol"].(string))
		ACLRule.Source = source
		ACLRule.SrcPortLow = uint32(rawRule["src_port_low"].(int))
		ACLRule.SrcPortHigh = uint32(rawRule["src_port_high"].(int))
		ACLRule.Destination = destination
		ACLRule.DstPortLow = uint32(rawRule["dst_port_low"].(int))
		ACLRule.DstPortHigh = uint32(rawRule["dst_port_high"].(int))
		ACLRule.Action = vpc.Action(rawRule["action"].(string))
		ACLRule.Description = types.ExpandStringPtr(rawRule["description"].(string))

		rules = append(rules, ACLRule)
	}

	return rules, nil
}

func flattenACLRules(rules []*vpc.ACLRule) any {
	if rules == nil {
		return nil
	}

	flattenedRules := []map[string]any(nil)

	var ruleScopeRegex = regexp.MustCompile(`^\(Rule scope: [^)]+\)\s*`)

	for _, rule := range rules {
		flattenedSource, err := types.FlattenIPNet(rule.Source)
		if err != nil {
			return nil
		}

		flattenedDestination, err := types.FlattenIPNet(rule.Destination)
		if err != nil {
			return nil
		}

		rawDescription := types.FlattenStringPtr(rule.Description)
		cleanDescription := ruleScopeRegex.ReplaceAllString(rawDescription.(string), "")

		flattenedRules = append(flattenedRules, map[string]any{
			"protocol":      rule.Protocol.String(),
			"source":        flattenedSource,
			"src_port_low":  int(rule.SrcPortLow),
			"src_port_high": int(rule.SrcPortHigh),
			"destination":   flattenedDestination,
			"dst_port_low":  int(rule.DstPortLow),
			"dst_port_high": int(rule.DstPortHigh),
			"action":        rule.Action.String(),
			"description":   cleanDescription,
		})
	}

	return flattenedRules
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
