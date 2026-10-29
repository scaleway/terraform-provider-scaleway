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

func TestResolveID(t *testing.T) {
	clientWithDefault, err := scw.NewClient(scw.WithDefaultRegion(scw.RegionNlAms))
	require.NoError(t, err)

	clientWithoutDefault, err := scw.NewClient()
	require.NoError(t, err)

	testCases := []struct {
		name        string
		idValue     string
		regionValue string
		client      *scw.Client
		region      scw.Region
		id          string
		err         string
	}{
		{
			name:        "regional id preferred over region attribute",
			idValue:     "fr-par/my-id",
			regionValue: "nl-ams",
			client:      clientWithDefault,
			region:      scw.RegionFrPar,
			id:          "my-id",
		},
		{
			name:        "bare uuid with region attribute",
			idValue:     "my-id",
			regionValue: "fr-par",
			client:      clientWithoutDefault,
			region:      scw.RegionFrPar,
			id:          "my-id",
		},
		{
			name:    "bare uuid with client default region",
			idValue: "my-id",
			client:  clientWithDefault,
			region:  scw.RegionNlAms,
			id:      "my-id",
		},
		{
			name:    "empty id",
			idValue: "",
			client:  clientWithDefault,
			err:     "id is empty",
		},
		{
			name:    "bare uuid without region",
			idValue: "my-id",
			client:  clientWithoutDefault,
			err:     regional.ErrRegionNotFound.Error(),
		},
		{
			name:    "invalid regional id",
			idValue: "not-a-region/my-id",
			client:  clientWithDefault,
			err:     `invalid regional id "not-a-region/my-id"`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			region, id, resolveErr := regional.ResolveID(tc.idValue, tc.regionValue, tc.client)
			if tc.err != "" {
				require.ErrorContains(t, resolveErr, tc.err)

				return
			}

			require.NoError(t, resolveErr)
			assert.Equal(t, tc.region, region)
			assert.Equal(t, tc.id, id)
		})
	}
}
