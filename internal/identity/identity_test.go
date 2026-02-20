package identity_test

import (
	"testing"

	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	"github.com/stretchr/testify/assert"
)

func TestParseMultiPartID(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		expected map[string]string
		keyOrder []string
	}{
		{
			name:     "two parts",
			id:       "11111111-1111-1111-1111-111111111111/plan-name",
			keyOrder: []string{"project_id", "name"},
			expected: map[string]string{
				"project_id": "11111111-1111-1111-1111-111111111111",
				"name":       "plan-name",
			},
		},
		{
			name:     "three parts",
			id:       "fr-par/11111111-1111-1111-1111-111111111111/resource",
			keyOrder: []string{"region", "project_id", "resource_id"},
			expected: map[string]string{
				"region":      "fr-par",
				"project_id":  "11111111-1111-1111-1111-111111111111",
				"resource_id": "resource",
			},
		},
		{
			name:     "single part",
			id:       "11111111-1111-1111-1111-111111111111",
			keyOrder: []string{"id"},
			expected: map[string]string{
				"id": "11111111-1111-1111-1111-111111111111",
			},
		},
		{
			name:     "more keys than parts, missing parts get empty string",
			id:       "only-one-part",
			keyOrder: []string{"first", "second", "third"},
			expected: map[string]string{
				"first": "only-one-part",
			},
		},
		{
			name:     "empty id",
			id:       "",
			keyOrder: []string{"project_id", "name"},
			expected: map[string]string{
				"project_id": "",
			},
		},
		{
			name:     "no keys",
			id:       "11111111-1111-1111-1111-111111111111",
			keyOrder: []string{},
			expected: map[string]string{},
		},
		{
			name:     "id with multiple slashes, last part contains remaining slashes",
			id:       "project/name/with/slashes",
			keyOrder: []string{"project_id", "name"},
			expected: map[string]string{
				"project_id": "project",
				"name":       "name/with/slashes",
			},
		},
		{
			name:     "uuid format parts",
			id:       "11111111-1111-1111-1111-111111111111/22222222-2222-2222-2222-222222222222",
			keyOrder: []string{"project_id", "resource_id"},
			expected: map[string]string{
				"project_id":  "11111111-1111-1111-1111-111111111111",
				"resource_id": "22222222-2222-2222-2222-222222222222",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := identity.ParseMultiPartID(tt.id, tt.keyOrder...)
			assert.Equal(t, tt.expected, result)
		})
	}
}
