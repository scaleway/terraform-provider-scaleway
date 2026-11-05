package baremetal

import (
	"testing"

	"github.com/scaleway/scaleway-sdk-go/api/baremetal/v1"
	"github.com/stretchr/testify/assert"
)

func TestIsPrivateNetworkServerOption(t *testing.T) {
	tests := []struct {
		option   *baremetal.ServerOption
		name     string
		expected bool
	}{
		{
			name:     "private network option",
			option:   &baremetal.ServerOption{ID: "cd4158d7-2d65-49be-8803-c4b8ab6f760c"},
			expected: true,
		},
		{
			name:     "10Gbps private network option",
			option:   &baremetal.ServerOption{ID: "0f467976-4c7b-4c8a-a26f-009a0196da49"},
			expected: true,
		},
		{
			name:     "25Gbps private network option",
			option:   &baremetal.ServerOption{ID: "9b4a83cc-4624-49fc-906a-a55ab4baf322"},
			expected: true,
		},
		{
			name:     "unknown ID with a private network name",
			option:   &baremetal.ServerOption{ID: "11111111-1111-1111-1111-111111111111", Name: "Private Network 100Gbps"},
			expected: true,
		},
		{
			name:     "other option",
			option:   &baremetal.ServerOption{ID: "931df052-d713-4674-8b58-96a63244c8e2", Name: "Remote Access"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, isPrivateNetworkServerOption(tt.option))
		})
	}
}
