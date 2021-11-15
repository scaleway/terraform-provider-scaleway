package scaleway

import (
	"testing"

	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/stretchr/testify/assert"
)

func TestIsEqualPrivateNetwork(t *testing.T) {
	tests := []struct {
		name     string
		A        *lb.PrivateNetwork
		B        *lb.PrivateNetwork
		expected bool
	}{
		{
			name:     "isEqualDHCP",
			A:        &lb.PrivateNetwork{PrivateNetworkID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8", DHCPConfig: &lb.PrivateNetworkDHCPConfig{}},
			B:        &lb.PrivateNetwork{PrivateNetworkID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8", DHCPConfig: &lb.PrivateNetworkDHCPConfig{}},
			expected: true,
		},
		{
			name:     "isEqualStatic",
			A:        &lb.PrivateNetwork{PrivateNetworkID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8", StaticConfig: &lb.PrivateNetworkStaticConfig{IPAddress: []string{"172.16.0.100", "172.16.0.101"}}},
			B:        &lb.PrivateNetwork{PrivateNetworkID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8", StaticConfig: &lb.PrivateNetworkStaticConfig{IPAddress: []string{"172.16.0.100", "172.16.0.101"}}},
			expected: true,
		},
		{
			name:     "areNotEqualStatic",
			A:        &lb.PrivateNetwork{PrivateNetworkID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8", StaticConfig: &lb.PrivateNetworkStaticConfig{IPAddress: []string{"172.16.0.100", "172.16.0.101"}}},
			B:        &lb.PrivateNetwork{PrivateNetworkID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8", StaticConfig: &lb.PrivateNetworkStaticConfig{IPAddress: []string{"172.16.0.101", "172.16.0.101"}}},
			expected: false,
		},
		{
			name:     "areNotEqualDHCPToStatic",
			A:        &lb.PrivateNetwork{PrivateNetworkID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8", DHCPConfig: &lb.PrivateNetworkDHCPConfig{}},
			B:        &lb.PrivateNetwork{PrivateNetworkID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8", StaticConfig: &lb.PrivateNetworkStaticConfig{IPAddress: []string{"172.16.0.101", "172.16.0.101"}}},
			expected: false,
		},
		{
			name:     "areNotEqualDHCPToStatic",
			A:        &lb.PrivateNetwork{PrivateNetworkID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8", StaticConfig: &lb.PrivateNetworkStaticConfig{IPAddress: []string{"172.16.0.101", "172.16.0.101"}}},
			B:        &lb.PrivateNetwork{PrivateNetworkID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8", DHCPConfig: &lb.PrivateNetworkDHCPConfig{}},
			expected: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, isPrivateNetworkEqual(tt.A, tt.B))
		})
	}
}
