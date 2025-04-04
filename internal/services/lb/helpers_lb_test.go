package lb_test

import (
	"testing"

	lbSDK "github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/lb"
	"github.com/stretchr/testify/assert"
)

func TestIsEqualPrivateNetwork(t *testing.T) {
	tests := []struct {
		A        *lbSDK.PrivateNetwork
		B        *lbSDK.PrivateNetwork
		name     string
		expected bool
	}{
		{
			name:     "isEqualDHCP",
			A:        &lbSDK.PrivateNetwork{PrivateNetworkID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8", DHCPConfig: &lbSDK.PrivateNetworkDHCPConfig{}},
			B:        &lbSDK.PrivateNetwork{PrivateNetworkID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8", DHCPConfig: &lbSDK.PrivateNetworkDHCPConfig{}},
			expected: true,
		},
		{
			name:     "isEqualStatic",
			A:        &lbSDK.PrivateNetwork{PrivateNetworkID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8", StaticConfig: &lbSDK.PrivateNetworkStaticConfig{IPAddress: scw.StringsPtr([]string{"172.16.0.100", "172.16.0.101"})}},
			B:        &lbSDK.PrivateNetwork{PrivateNetworkID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8", StaticConfig: &lbSDK.PrivateNetworkStaticConfig{IPAddress: scw.StringsPtr([]string{"172.16.0.100", "172.16.0.101"})}},
			expected: true,
		},
		{
			name:     "areNotEqualStatic",
			A:        &lbSDK.PrivateNetwork{PrivateNetworkID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8", StaticConfig: &lbSDK.PrivateNetworkStaticConfig{IPAddress: scw.StringsPtr([]string{"172.16.0.100", "172.16.0.101"})}},
			B:        &lbSDK.PrivateNetwork{PrivateNetworkID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8", StaticConfig: &lbSDK.PrivateNetworkStaticConfig{IPAddress: scw.StringsPtr([]string{"172.16.0.101", "172.16.0.101"})}},
			expected: false,
		},
		{
			name:     "areNotEqualDHCPToStatic",
			A:        &lbSDK.PrivateNetwork{PrivateNetworkID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8", DHCPConfig: &lbSDK.PrivateNetworkDHCPConfig{}},
			B:        &lbSDK.PrivateNetwork{PrivateNetworkID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8", StaticConfig: &lbSDK.PrivateNetworkStaticConfig{IPAddress: scw.StringsPtr([]string{"172.16.0.101", "172.16.0.101"})}},
			expected: false,
		},
		{
			name:     "areNotEqualDHCPToStatic",
			A:        &lbSDK.PrivateNetwork{PrivateNetworkID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8", StaticConfig: &lbSDK.PrivateNetworkStaticConfig{IPAddress: scw.StringsPtr([]string{"172.16.0.101", "172.16.0.101"})}},
			B:        &lbSDK.PrivateNetwork{PrivateNetworkID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8", DHCPConfig: &lbSDK.PrivateNetworkDHCPConfig{}},
			expected: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, lb.IsPrivateNetworkEqual(tt.A, tt.B))
		})
	}
}

func TestPrivateNetworksCompare(t *testing.T) {
	tests := []struct {
		name             string
		oldPNs           []*lbSDK.PrivateNetwork
		newPNs           []*lbSDK.PrivateNetwork
		expectedToDetach []*lbSDK.PrivateNetwork
		expectedToAttach []*lbSDK.PrivateNetwork
	}{
		{
			name: "no changes",
			oldPNs: []*lbSDK.PrivateNetwork{
				{PrivateNetworkID: "pn1", StaticConfig: &lbSDK.PrivateNetworkStaticConfig{IPAddress: scw.StringsPtr([]string{"192.168.1.1"})}},
				{PrivateNetworkID: "pn2", DHCPConfig: &lbSDK.PrivateNetworkDHCPConfig{}},
			},
			newPNs: []*lbSDK.PrivateNetwork{
				{PrivateNetworkID: "pn1", StaticConfig: &lbSDK.PrivateNetworkStaticConfig{IPAddress: scw.StringsPtr([]string{"192.168.1.1"})}},
				{PrivateNetworkID: "pn2", DHCPConfig: &lbSDK.PrivateNetworkDHCPConfig{}},
			},
			expectedToDetach: []*lbSDK.PrivateNetwork(nil),
			expectedToAttach: []*lbSDK.PrivateNetwork(nil),
		},
		{
			name: "private network removed",
			oldPNs: []*lbSDK.PrivateNetwork{
				{PrivateNetworkID: "pn1", StaticConfig: &lbSDK.PrivateNetworkStaticConfig{IPAddress: scw.StringsPtr([]string{"192.168.1.1"})}},
				{PrivateNetworkID: "pn2", DHCPConfig: &lbSDK.PrivateNetworkDHCPConfig{}},
			},
			newPNs: []*lbSDK.PrivateNetwork{
				{PrivateNetworkID: "pn1", StaticConfig: &lbSDK.PrivateNetworkStaticConfig{IPAddress: scw.StringsPtr([]string{"192.168.1.1"})}},
			},
			expectedToDetach: []*lbSDK.PrivateNetwork{
				{PrivateNetworkID: "pn2", DHCPConfig: &lbSDK.PrivateNetworkDHCPConfig{}},
			},
			expectedToAttach: []*lbSDK.PrivateNetwork(nil),
		},
		{
			name: "private network added",
			oldPNs: []*lbSDK.PrivateNetwork{
				{PrivateNetworkID: "pn1", StaticConfig: &lbSDK.PrivateNetworkStaticConfig{IPAddress: scw.StringsPtr([]string{"192.168.1.1"})}},
			},
			newPNs: []*lbSDK.PrivateNetwork{
				{PrivateNetworkID: "pn1", StaticConfig: &lbSDK.PrivateNetworkStaticConfig{IPAddress: scw.StringsPtr([]string{"192.168.1.1"})}},
				{PrivateNetworkID: "pn2", DHCPConfig: &lbSDK.PrivateNetworkDHCPConfig{}},
			},
			expectedToDetach: []*lbSDK.PrivateNetwork(nil),
			expectedToAttach: []*lbSDK.PrivateNetwork{
				{PrivateNetworkID: "pn2", DHCPConfig: &lbSDK.PrivateNetworkDHCPConfig{}},
			},
		},
		{
			name: "private network static configuration changed",
			oldPNs: []*lbSDK.PrivateNetwork{
				{PrivateNetworkID: "pn1", StaticConfig: &lbSDK.PrivateNetworkStaticConfig{IPAddress: scw.StringsPtr([]string{"192.168.1.1"})}},
			},
			newPNs: []*lbSDK.PrivateNetwork{
				{PrivateNetworkID: "pn1", StaticConfig: &lbSDK.PrivateNetworkStaticConfig{IPAddress: scw.StringsPtr([]string{"192.168.1.2"})}},
			},
			expectedToDetach: []*lbSDK.PrivateNetwork{
				{PrivateNetworkID: "pn1", StaticConfig: &lbSDK.PrivateNetworkStaticConfig{IPAddress: scw.StringsPtr([]string{"192.168.1.1"})}},
			},
			expectedToAttach: []*lbSDK.PrivateNetwork{
				{PrivateNetworkID: "pn1", StaticConfig: &lbSDK.PrivateNetworkStaticConfig{IPAddress: scw.StringsPtr([]string{"192.168.1.2"})}},
			},
		},
		{
			name: "private network configuration changed from static to DHCP",
			oldPNs: []*lbSDK.PrivateNetwork{
				{PrivateNetworkID: "pn1", StaticConfig: &lbSDK.PrivateNetworkStaticConfig{IPAddress: scw.StringsPtr([]string{"192.168.1.1"})}},
			},
			newPNs: []*lbSDK.PrivateNetwork{
				{PrivateNetworkID: "pn1", DHCPConfig: &lbSDK.PrivateNetworkDHCPConfig{}},
			},
			expectedToDetach: []*lbSDK.PrivateNetwork{
				{PrivateNetworkID: "pn1", StaticConfig: &lbSDK.PrivateNetworkStaticConfig{IPAddress: scw.StringsPtr([]string{"192.168.1.1"})}},
			},
			expectedToAttach: []*lbSDK.PrivateNetwork{
				{PrivateNetworkID: "pn1", DHCPConfig: &lbSDK.PrivateNetworkDHCPConfig{}},
			},
		},
		{
			name: "multiple private networks removed",
			oldPNs: []*lbSDK.PrivateNetwork{
				{PrivateNetworkID: "pn1", StaticConfig: &lbSDK.PrivateNetworkStaticConfig{IPAddress: scw.StringsPtr([]string{"192.168.1.1"})}},
				{PrivateNetworkID: "pn2", DHCPConfig: &lbSDK.PrivateNetworkDHCPConfig{}},
				{PrivateNetworkID: "pn3", StaticConfig: &lbSDK.PrivateNetworkStaticConfig{IPAddress: scw.StringsPtr([]string{"192.168.1.3"})}},
			},
			newPNs: []*lbSDK.PrivateNetwork{
				{PrivateNetworkID: "pn1", StaticConfig: &lbSDK.PrivateNetworkStaticConfig{IPAddress: scw.StringsPtr([]string{"192.168.1.1"})}},
			},
			expectedToDetach: []*lbSDK.PrivateNetwork{
				{PrivateNetworkID: "pn2", DHCPConfig: &lbSDK.PrivateNetworkDHCPConfig{}},
				{PrivateNetworkID: "pn3", StaticConfig: &lbSDK.PrivateNetworkStaticConfig{IPAddress: scw.StringsPtr([]string{"192.168.1.3"})}},
			},
			expectedToAttach: []*lbSDK.PrivateNetwork(nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toDetach, toAttach := lb.PrivateNetworksCompare(tt.oldPNs, tt.newPNs)
			assert.ElementsMatch(t, tt.expectedToDetach, toDetach)
			assert.ElementsMatch(t, tt.expectedToAttach, toAttach)
		})
	}
}
