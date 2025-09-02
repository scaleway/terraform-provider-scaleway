package verify

import (
	"net"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func IsStandaloneIPorCIDR() schema.SchemaValidateDiagFunc {
	return func(value any, path cty.Path) diag.Diagnostics {
		ip, isString := value.(string)
		if !isString {
			return diag.Diagnostics{diag.Diagnostic{
				Severity:      diag.Error,
				AttributePath: path,
				Summary:       "invalid, input not a string: " + ip,
			}}
		}

		// Check if it's a standalone IP address
		if net.ParseIP(ip) != nil {
			return nil
		}

		// Check if it's an IP with CIDR notation
		_, _, err := net.ParseCIDR(ip)
		if err != nil {
			return diag.Diagnostics{diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       "neither a valid IP address or CIDR notation: " + ip,
				AttributePath: path,
			}}
		}

		return nil
	}
}
