package types

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"

	"github.com/scaleway/scaleway-sdk-go/scw"
)

// NetIPNil define the nil string return by (*net.IP).String()
const NetIPNil = "<nil>"

func ExpandIPNet(raw string) (scw.IPNet, error) {
	if raw == "" {
		return scw.IPNet{}, nil
	}

	var ipNet scw.IPNet

	err := json.Unmarshal([]byte(strconv.Quote(raw)), &ipNet)
	if err != nil {
		return scw.IPNet{}, fmt.Errorf("%s could not be marshaled: %w", raw, err)
	}

	return ipNet, nil
}

func FlattenIPPtr(ip *net.IP) any {
	if ip == nil {
		return ""
	}

	return ip.String()
}

func FlattenIPNet(ipNet scw.IPNet) (string, error) {
	raw, err := json.Marshal(ipNet)
	if err != nil {
		return "", err
	}

	return string(raw[1 : len(raw)-1]), nil // remove quotes
}

// NormalizeIPToCIDR converts a standalone IP address to CIDR notation with default mask
// If the input is already in CIDR notation, it returns it unchanged
// IPv4 addresses get /32 mask, IPv6 addresses get /128 mask
func NormalizeIPToCIDR(raw string) (string, error) {
	if raw == "" {
		return "", nil
	}

	// Check if it's already a valid CIDR
	if _, _, err := net.ParseCIDR(raw); err == nil {
		return raw, nil
	}

	// Try to parse as standalone IP
	ip := net.ParseIP(raw)
	if ip == nil {
		return "", fmt.Errorf("invalid IP address or CIDR notation: %s", raw)
	}

	// Add default mask based on IP version
	if ip.To4() != nil {
		// IPv4 address
		return raw + "/32", nil
	}

	// IPv6 address
	return raw + "/128", nil
}
