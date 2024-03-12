package locality_test

import (
	"testing"

	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseLocalizedID(t *testing.T) {
	testCases := []struct {
		name       string
		localityID string
		id         string
		locality   string
		err        string
	}{
		{
			name:       "simple",
			localityID: "fr-par-1/my-id",
			id:         "my-id",
			locality:   "fr-par-1",
		},
		{
			name:       "id with a region",
			localityID: "fr-par/my-id",
			id:         "my-id",
			locality:   "fr-par",
		},
		{
			name:       "empty",
			localityID: "",
			err:        "cant parse localized id: ",
		},
		{
			name:       "without locality",
			localityID: "my-id",
			err:        "cant parse localized id: my-id",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			l, id, err := locality.ParseLocalizedID(tc.localityID)
			if tc.err != "" {
				require.EqualError(t, err, tc.err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.locality, l)
				assert.Equal(t, tc.id, id)
			}
		})
	}
}

func TestParseLocalizedNestedID(t *testing.T) {
	testCases := []struct {
		name       string
		localityID string
		innerID    string
		outerID    string
		locality   string
		err        string
	}{
		{
			name:       "id with a sub directory",
			localityID: "fr-par/my-id/subdir",
			innerID:    "my-id",
			outerID:    "subdir",
			locality:   "fr-par",
		},
		{
			name:       "id with multiple sub directories",
			localityID: "fr-par/my-id/subdir/foo/bar",
			innerID:    "my-id",
			outerID:    "subdir/foo/bar",
			locality:   "fr-par",
		},
		{
			name:       "simple",
			localityID: "fr-par-1/my-id",
			err:        "cant parse localized id: fr-par-1/my-id",
		},
		{
			name:       "empty",
			localityID: "",
			err:        "cant parse localized id: ",
		},
		{
			name:       "without locality",
			localityID: "my-id",
			err:        "cant parse localized id: my-id",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			l, innerID, outerID, err := locality.ParseLocalizedNestedID(tc.localityID)
			if tc.err != "" {
				require.EqualError(t, err, tc.err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.locality, l)
				assert.Equal(t, tc.innerID, innerID)
				assert.Equal(t, tc.outerID, outerID)
			}
		})
	}
}
