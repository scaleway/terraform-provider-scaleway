package scaleway

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
)

// securityGroupRuleExpand transform a state rule to an api one.
func securityGroupRuleExpand(i interface{}) (*instance.SecurityGroupRule, error) {
	rawRule := i.(map[string]interface{})

	portFrom, portTo := uint32(1), uint32(1)

	var portRange string

	if portRangeRaw, ok := rawRule["port_range"]; ok && len(portRangeRaw.(string)) > 0 {
		portRange = portRangeRaw.(string)
		_, _ = fmt.Sscanf(portRange, "%d-%d", &portFrom, &portTo)
	}

	action, _ := rawRule["action"].(string)

	if portTo == 1 {
		portTo = portFrom
	}

	editable, _ := rawRule["editable"].(bool)
	protocol := strings.ToLower(instance.SecurityGroupRuleProtocol(rawRule["protocol"].(string)).String())
	_, err := net.LookupPort(protocol, strconv.Itoa(int(portFrom)))
	if err != nil {
		return nil, fmt.Errorf("port from %d, with error: %w", portFrom, err)
	}

	_, err = net.LookupPort(protocol, strconv.Itoa(int(portTo)))
	if err != nil {
		return nil, fmt.Errorf("port to %d, with error: %w", portFrom, err)
	}

	ipNetRange, err := expandIPNet(rawRule["ip_range"].(string))
	if err != nil {
		return nil, err
	}
	rule := &instance.SecurityGroupRule{
		DestPortFrom: &portFrom,
		DestPortTo:   &portTo,
		Editable:     editable,
		Protocol:     instance.SecurityGroupRuleProtocol(rawRule["protocol"].(string)),
		IPRange:      ipNetRange,
		Action:       instance.SecurityGroupRuleAction(action),
	}

	// Handle when no port is specified. We allow all ports
	if portFrom == 1 && portTo == 1 {
		rule.DestPortFrom = nil
		rule.DestPortTo = nil
	}

	return rule, nil
}
