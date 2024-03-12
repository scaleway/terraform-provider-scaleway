package zonal_test

import (
	"testing"

	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewZonedId(t *testing.T) {
	assert.Equal(t, "fr-par-1/my-id", zonal.NewIDString(scw.ZoneFrPar1, "my-id"))
}

func TestParseZonedID(t *testing.T) {
	testCases := []struct {
		name       string
		localityID string
		id         string
		zone       scw.Zone
		err        string
	}{
		{
			name:       "simple",
			localityID: "fr-par-1/my-id",
			id:         "my-id",
			zone:       scw.ZoneFrPar1,
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
			zone, id, err := zonal.ParseID(tc.localityID)
			if tc.err != "" {
				require.EqualError(t, err, tc.err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.zone, zone)
				assert.Equal(t, tc.id, id)
			}
		})
	}
}
