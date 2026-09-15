package identity_test

import (
	"testing"

	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpgradeDefaultRegionalToComposite(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		partKeys []string
		rawState map[string]any
		want     map[string]any
		wantErr  string
	}{
		{
			name:     "database from default regional",
			partKeys: []string{"instance_id", "database_name"},
			rawState: map[string]any{
				"id":     "11111111-1111-1111-1111-111111111111/toto_db",
				"region": "fr-par",
			},
			want: map[string]any{
				"region":        "fr-par",
				"instance_id":   "11111111-1111-1111-1111-111111111111",
				"database_name": "toto_db",
			},
		},
		{
			name:     "user from default regional",
			partKeys: []string{"instance_id", "name"},
			rawState: map[string]any{
				"id":     "11111111-1111-1111-1111-111111111111/alice",
				"region": "nl-ams",
			},
			want: map[string]any{
				"region":      "nl-ams",
				"instance_id": "11111111-1111-1111-1111-111111111111",
				"name":        "alice",
			},
		},
		{
			name:     "privilege from default regional",
			partKeys: []string{"instance_id", "database_name", "user_name"},
			rawState: map[string]any{
				"id":     "11111111-1111-1111-1111-111111111111/toto_db/alice",
				"region": "fr-par",
			},
			want: map[string]any{
				"region":        "fr-par",
				"instance_id":   "11111111-1111-1111-1111-111111111111",
				"database_name": "toto_db",
				"user_name":     "alice",
			},
		},
		{
			name:     "already composite identity from 2.83.0 is passed through",
			partKeys: []string{"instance_id", "database_name"},
			rawState: map[string]any{
				"region":        "fr-par",
				"instance_id":   "11111111-1111-1111-1111-111111111111",
				"database_name": "toto_db",
			},
			want: map[string]any{
				"region":        "fr-par",
				"instance_id":   "11111111-1111-1111-1111-111111111111",
				"database_name": "toto_db",
			},
		},
		{
			name:     "nil state",
			partKeys: []string{"instance_id", "database_name"},
			rawState: nil,
			want:     map[string]any{},
		},
		{
			name:     "wrong number of parts",
			partKeys: []string{"instance_id", "database_name"},
			rawState: map[string]any{
				"id":     "only-one-part",
				"region": "fr-par",
			},
			wantErr: `identity id "only-one-part" does not have 2 parts`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := identity.UpgradeDefaultRegionalToComposite(tt.partKeys...)(t.Context(), tt.rawState, nil)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCompositeRegionalIdentityVersion(t *testing.T) {
	t.Parallel()

	id := identity.CompositeRegionalIdentity("instance_id", "database_name")
	require.NotNil(t, id)
	assert.Equal(t, int64(1), id.Version)
	require.Len(t, id.IdentityUpgraders, 1)
	assert.Equal(t, int64(0), id.IdentityUpgraders[0].Version)
}

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
