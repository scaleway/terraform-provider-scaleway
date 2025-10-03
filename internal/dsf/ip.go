package dsf

import (
	"net"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DiffSuppressFuncStandaloneIPandCIDR(_, oldValue, newValue string, _ *schema.ResourceData) bool {
	parseIPOrCIDR := func(s string) net.IP {
		if ip, _, err := net.ParseCIDR(s); err == nil {
			return ip
		}

		return net.ParseIP(s)
	}

	oldIP := parseIPOrCIDR(oldValue)
	newIP := parseIPOrCIDR(newValue)

	return oldIP != nil && newIP != nil && oldIP.Equal(newIP)
}
