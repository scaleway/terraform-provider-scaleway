package verify

import (
	"fmt"
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

// IsIPv4CIDR validates that the value is a valid IPv4 address or range in CIDR notation.
// IPv6 is not supported by some Scaleway APIs (e.g. RDB ACL, Redis ACL).
func IsIPv4CIDR() schema.SchemaValidateDiagFunc {
	return func(i any, path cty.Path) diag.Diagnostics {
		v, ok := i.(string)
		if !ok {
			return diag.Diagnostics{diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       "expected type to be string",
				AttributePath: path,
			}}
		}

		ip, _, err := net.ParseCIDR(v)
		if err != nil {
			return diag.Diagnostics{diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       fmt.Sprintf("expected a valid CIDR, got %v", v),
				AttributePath: path,
			}}
		}

		if ip.To4() == nil {
			return diag.Diagnostics{diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       "must be an IPv4 address or range in CIDR notation (IPv6 is not supported by the Scaleway API)",
				AttributePath: path,
			}}
		}

		return nil
	}
}
