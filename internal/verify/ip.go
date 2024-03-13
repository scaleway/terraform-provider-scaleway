package verify

import (
	"fmt"
	"net"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func IsStandaloneIPorCIDR() schema.SchemaValidateFunc {
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
		_, _, err := net.ParseCIDR(ip)
		if err != nil {
			errs = append(errs, fmt.Errorf("%q is not a valid IP address or CIDR notation: %s", key, ip))
		}

		return
	}
}
