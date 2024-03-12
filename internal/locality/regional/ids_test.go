package regional_test

import (
	"testing"

	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRegionalId(t *testing.T) {
	assert.Equal(t, "fr-par/my-id", regional.NewIDString(scw.RegionFrPar, "my-id"))
}

func TestParseRegionID(t *testing.T) {
	testCases := []struct {
		name       string
		localityID string
		id         string
		region     scw.Region
		err        string
	}{
		{
			name:       "simple",
			localityID: "fr-par/my-id",
			id:         "my-id",
			region:     scw.RegionFrPar,
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
			region, id, err := regional.ParseID(tc.localityID)
			if tc.err != "" {
				require.EqualError(t, err, tc.err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.region, region)
				assert.Equal(t, tc.id, id)
			}
		})
	}
}
