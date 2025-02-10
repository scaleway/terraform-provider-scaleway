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

func FlattenIPPtr(ip *net.IP) interface{} {
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
