package instance_test

import (
	"testing"

	instanceSDK "github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/instance"
	"github.com/stretchr/testify/require"
)

func rawRule(port, priority int, action string) map[string]any {
	return map[string]any{
		"action":     action,
		"protocol":   instanceSDK.SecurityGroupRuleProtocolTCP.String(),
		"port":       port,
		"port_range": "",
		"ip_range":   "0.0.0.0/0",
		"ip":         "",
		"priority":   priority,
	}
}

func TestBuildSetSecurityGroupRulesRequest_SortsByPriority(t *testing.T) {
	zone := scw.ZoneFrPar1

	stateRules := map[instanceSDK.SecurityGroupRuleDirection][]any{
		// Rules are intentionally not ordered by priority in state to check the sorting.
		instanceSDK.SecurityGroupRuleDirectionInbound: {
			rawRule(80, 20, "accept"),
			rawRule(22, 10, "accept"),
		},
		instanceSDK.SecurityGroupRuleDirectionOutbound: {
			rawRule(25, 40, "drop"),
			rawRule(443, 30, "drop"),
		},
	}

	requestRules, err := instance.BuildSetSecurityGroupRulesRequest(stateRules, zone)
	require.NoError(t, err)
	require.Len(t, requestRules, 4)

	byDirection := map[instanceSDK.SecurityGroupRuleDirection][]*instanceSDK.SetSecurityGroupRulesRequestRule{}
	for _, rule := range requestRules {
		byDirection[rule.Direction] = append(byDirection[rule.Direction], rule)
	}

	// Inbound rules must be sorted by ascending priority (22 before 80).
	inbound := byDirection[instanceSDK.SecurityGroupRuleDirectionInbound]
	require.Len(t, inbound, 2)
	require.Equal(t, uint32(10), inbound[0].Position)
	require.Equal(t, uint32(22), *inbound[0].DestPortFrom)
	require.Equal(t, uint32(20), inbound[1].Position)
	require.Equal(t, uint32(80), *inbound[1].DestPortFrom)

	// Outbound rules must be sorted by ascending priority (443 before 25).
	outbound := byDirection[instanceSDK.SecurityGroupRuleDirectionOutbound]
	require.Len(t, outbound, 2)
	require.Equal(t, uint32(30), outbound[0].Position)
	require.Equal(t, uint32(443), *outbound[0].DestPortFrom)
	require.Equal(t, uint32(40), outbound[1].Position)
	require.Equal(t, uint32(25), *outbound[1].DestPortFrom)
}

func TestBuildSetSecurityGroupRulesRequest_KeepsOrderForEqualPriorities(t *testing.T) {
	zone := scw.ZoneFrPar1

	// All rules share the same priority (0): their configuration order must be kept.
	stateRules := map[instanceSDK.SecurityGroupRuleDirection][]any{
		instanceSDK.SecurityGroupRuleDirectionInbound: {
			rawRule(80, 0, "accept"),
			rawRule(443, 0, "drop"),
		},
	}

	requestRules, err := instance.BuildSetSecurityGroupRulesRequest(stateRules, zone)
	require.NoError(t, err)
	require.Len(t, requestRules, 2)
	require.Equal(t, uint32(80), *requestRules[0].DestPortFrom)
	require.Equal(t, uint32(443), *requestRules[1].DestPortFrom)
	require.Equal(t, uint32(0), requestRules[0].Position)
	require.Equal(t, uint32(0), requestRules[1].Position)
}

func TestBuildSetSecurityGroupRulesRequest_DefaultPriorityIsZero(t *testing.T) {
	zone := scw.ZoneFrPar1

	rule := rawRule(80, 0, "accept")
	delete(rule, "priority")

	stateRules := map[instanceSDK.SecurityGroupRuleDirection][]any{
		instanceSDK.SecurityGroupRuleDirectionInbound: {rule},
	}

	requestRules, err := instance.BuildSetSecurityGroupRulesRequest(stateRules, zone)
	require.NoError(t, err)
	require.Len(t, requestRules, 1)
	require.Equal(t, uint32(0), requestRules[0].Position)
}
