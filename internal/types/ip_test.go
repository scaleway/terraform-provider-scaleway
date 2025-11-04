package types_test

import (
	"testing"

	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeIPToCIDR(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		{
			name:        "IPv4 address without mask",
			input:       "10.213.254.3",
			expected:    "10.213.254.3/32",
			expectError: false,
		},
		{
			name:        "IPv4 address with /32 mask",
			input:       "10.213.254.3/32",
			expected:    "10.213.254.3/32",
			expectError: false,
		},
		{
			name:        "IPv4 address with /24 mask",
			input:       "192.168.1.0/24",
			expected:    "192.168.1.0/24",
			expectError: false,
		},
		{
			name:        "IPv4 address with /16 mask",
			input:       "10.0.0.0/16",
			expected:    "10.0.0.0/16",
			expectError: false,
		},
		{
			name:        "IPv6 address without mask",
			input:       "2001:db8::1",
			expected:    "2001:db8::1/128",
			expectError: false,
		},
		{
			name:        "IPv6 address with /128 mask",
			input:       "2001:db8::1/128",
			expected:    "2001:db8::1/128",
			expectError: false,
		},
		{
			name:        "IPv6 address with /64 mask",
			input:       "2001:db8::/64",
			expected:    "2001:db8::/64",
			expectError: false,
		},
		{
			name:        "empty string",
			input:       "",
			expected:    "",
			expectError: false,
		},
		{
			name:        "invalid IP",
			input:       "not-an-ip",
			expected:    "",
			expectError: true,
		},
		{
			name:        "invalid CIDR",
			input:       "10.0.0.1/33",
			expected:    "",
			expectError: true,
		},
		{
			name:        "localhost IPv4",
			input:       "127.0.0.1",
			expected:    "127.0.0.1/32",
			expectError: false,
		},
		{
			name:        "localhost IPv6",
			input:       "::1",
			expected:    "::1/128",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := types.NormalizeIPToCIDR(tt.input)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

