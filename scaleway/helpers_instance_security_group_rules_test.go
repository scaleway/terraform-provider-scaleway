package scaleway

import (
	"testing"

	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseIPRange(t *testing.T) {
	expectedIP1, _ := expandIPNet("0.0.0.0/32")
	expectedIP2, _ := expandIPNet("195.1.3.4/32")

	toUint32Prt := func(v uint32) *uint32 {
		return &v
	}
	testCases := []struct {
		name         string
		ruleRaw      map[string]interface{}
		expectedRule *instance.SecurityGroupRule
		err          string
	}{
		{
			name: "simple",
			ruleRaw: map[string]interface{}{
				"action":     instance.SecurityGroupRuleActionAccept.String(),
				"ip_range":   "0.0.0.0/32",
				"editable":   true,
				"protocol":   instance.SecurityGroupRuleProtocolTCP.String(),
				"port_range": "80-80",
			},
			expectedRule: &instance.SecurityGroupRule{
				Action:       instance.SecurityGroupRuleActionAccept,
				Protocol:     instance.SecurityGroupRuleProtocolTCP,
				IPRange:      expectedIP1,
				Editable:     true,
				DestPortFrom: toUint32Prt(80),
				DestPortTo:   toUint32Prt(80),
			},
		},
		{
			name: "simple",
			ruleRaw: map[string]interface{}{
				"action":     instance.SecurityGroupRuleActionDrop.String(),
				"ip_range":   "195.1.3.4/32",
				"editable":   true,
				"protocol":   instance.SecurityGroupRuleProtocolTCP.String(),
				"port_range": "22",
			},
			expectedRule: &instance.SecurityGroupRule{
				Action:       instance.SecurityGroupRuleActionDrop,
				Protocol:     instance.SecurityGroupRuleProtocolTCP,
				IPRange:      expectedIP2,
				Editable:     true,
				DestPortFrom: toUint32Prt(22),
				DestPortTo:   toUint32Prt(22),
			},
		},
		{
			name: "simple",
			ruleRaw: map[string]interface{}{
				"action":     instance.SecurityGroupRuleActionDrop.String(),
				"ip_range":   "195.1.3.4/32",
				"editable":   true,
				"protocol":   instance.SecurityGroupRuleProtocolTCP.String(),
				"port_range": "6500000",
			},
			err: "port from 6500000, with error: address 6500000: invalid port",
		},
		{
			name: "simple",
			ruleRaw: map[string]interface{}{
				"action":   instance.SecurityGroupRuleActionDrop.String(),
				"ip_range": "195.1.3.4/32",
				"editable": true,
				"protocol": instance.SecurityGroupRuleProtocolTCP.String(),
			},
			expectedRule: &instance.SecurityGroupRule{
				Action:       instance.SecurityGroupRuleActionDrop,
				Protocol:     instance.SecurityGroupRuleProtocolTCP,
				IPRange:      expectedIP2,
				Editable:     true,
				DestPortFrom: nil,
				DestPortTo:   nil,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rule, err := securityGroupRuleExpand(tc.ruleRaw)
			if tc.err != "" {
				require.EqualError(t, err, tc.err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedRule, rule)
			}
		})
	}
}
