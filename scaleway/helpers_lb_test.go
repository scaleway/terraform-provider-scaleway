package scaleway

import (
	"testing"

	lbSDK "github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/stretchr/testify/assert"
)

func TestIsEqualPrivateNetwork(t *testing.T) {
	tests := []struct {
		name     string
		A        *lbSDK.PrivateNetwork
		B        *lbSDK.PrivateNetwork
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
			A:        &lbSDK.PrivateNetwork{PrivateNetworkID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8", StaticConfig: &lbSDK.PrivateNetworkStaticConfig{IPAddress: []string{"172.16.0.100", "172.16.0.101"}}},
			B:        &lbSDK.PrivateNetwork{PrivateNetworkID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8", StaticConfig: &lbSDK.PrivateNetworkStaticConfig{IPAddress: []string{"172.16.0.100", "172.16.0.101"}}},
			expected: true,
		},
		{
			name:     "areNotEqualStatic",
			A:        &lbSDK.PrivateNetwork{PrivateNetworkID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8", StaticConfig: &lbSDK.PrivateNetworkStaticConfig{IPAddress: []string{"172.16.0.100", "172.16.0.101"}}},
			B:        &lbSDK.PrivateNetwork{PrivateNetworkID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8", StaticConfig: &lbSDK.PrivateNetworkStaticConfig{IPAddress: []string{"172.16.0.101", "172.16.0.101"}}},
			expected: false,
		},
		{
			name:     "areNotEqualDHCPToStatic",
			A:        &lbSDK.PrivateNetwork{PrivateNetworkID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8", DHCPConfig: &lbSDK.PrivateNetworkDHCPConfig{}},
			B:        &lbSDK.PrivateNetwork{PrivateNetworkID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8", StaticConfig: &lbSDK.PrivateNetworkStaticConfig{IPAddress: []string{"172.16.0.101", "172.16.0.101"}}},
			expected: false,
		},
		{
			name:     "areNotEqualDHCPToStatic",
			A:        &lbSDK.PrivateNetwork{PrivateNetworkID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8", StaticConfig: &lbSDK.PrivateNetworkStaticConfig{IPAddress: []string{"172.16.0.101", "172.16.0.101"}}},
			B:        &lbSDK.PrivateNetwork{PrivateNetworkID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8", DHCPConfig: &lbSDK.PrivateNetworkDHCPConfig{}},
			expected: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, isPrivateNetworkEqual(tt.A, tt.B))
		})
	}
}
