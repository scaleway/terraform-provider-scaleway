package scaleway

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/scaleway-sdk-go/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseLocalizedID(t *testing.T) {

	testCases := []struct {
		name       string
		localityId string
		id         string
		locality   string
		err        string
	}{
		{
			name:       "simple",
			localityId: "fr-par-1/my-id",
			id:         "my-id",
			locality:   "fr-par-1",
		},
		{
			name:       "id with slashed",
			localityId: "fr-par-1/my/id",
			id:         "my/id",
			locality:   "fr-par-1",
		},
		{
			name:       "id with a region",
			localityId: "fr-par/my-id",
			id:         "my-id",
			locality:   "fr-par",
		},
		{
			name:       "empty",
			localityId: "",
			err:        "cant parse localized id: ",
		},
		{
			name:       "without locality",
			localityId: "my-id",
			err:        "cant parse localized id: my-id",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			locality, id, err := ParseLocalizedID(testCase.localityId)
			if testCase.err != "" {
				require.EqualError(t, err, testCase.err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.locality, locality)
				assert.Equal(t, testCase.id, id)
			}
		})
	}

}

func TestParseZonedID(t *testing.T) {

	testCases := []struct {
		name       string
		localityId string
		id         string
		zone       utils.Zone
		err        string
	}{
		{
			name:       "simple",
			localityId: "fr-par-1/my-id",
			id:         "my-id",
			zone:       utils.ZoneFrPar1,
		},
		{
			name:       "id with slashed",
			localityId: "fr-par-1/my/id",
			id:         "my/id",
			zone:       utils.ZoneFrPar1,
		},
		{
			name:       "id with a region",
			localityId: "fr-par/my-id",
			id:         "my-id",
			zone:       utils.Zone("fr-par"),
		},
		{
			name:       "empty",
			localityId: "",
			err:        "cant parse localized id: ",
		},
		{
			name:       "without locality",
			localityId: "my-id",
			err:        "cant parse localized id: my-id",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			zone, id, err := ParseZonedID(testCase.localityId)
			if testCase.err != "" {
				require.EqualError(t, err, testCase.err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.zone, zone)
				assert.Equal(t, testCase.id, id)
			}
		})
	}

}

func TestParseRegionID(t *testing.T) {

	testCases := []struct {
		name       string
		localityId string
		id         string
		region     utils.Region
		err        string
	}{
		{
			name:       "simple",
			localityId: "fr-par/my-id",
			id:         "my-id",
			region:     utils.RegionFrPar,
		},
		{
			name:       "id with slashed",
			localityId: "fr-par/my/id",
			id:         "my/id",
			region:     utils.RegionFrPar,
		},
		{
			name:       "empty",
			localityId: "",
			err:        "cant parse localized id: ",
		},
		{
			name:       "without locality",
			localityId: "my-id",
			err:        "cant parse localized id: my-id",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			region, id, err := ParseRegionalID(testCase.localityId)
			if testCase.err != "" {
				require.EqualError(t, err, testCase.err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.region, region)
				assert.Equal(t, testCase.id, id)
			}
		})
	}

}

func TestNewZonedId(t *testing.T) {
	assert.Equal(t, "fr-par-1/my-id", NewZonedId(utils.ZoneFrPar1, "my-id"))
}

func TestNewRegionalId(t *testing.T) {
	assert.Equal(t, "fr-par/my-id", NewRegionalId(utils.RegionFrPar, "my-id"))
}

func TestIsHTTPCodeError(t *testing.T) {
	assert.True(t, IsHTTPCodeError(&scw.ResponseError{StatusCode: http.StatusBadRequest}, http.StatusBadRequest))
	assert.False(t, IsHTTPCodeError(nil, http.StatusBadRequest))
	assert.False(t, IsHTTPCodeError(&scw.ResponseError{StatusCode: http.StatusBadRequest}, http.StatusNotFound))
	assert.False(t, IsHTTPCodeError(fmt.Errorf("not an http error"), http.StatusNotFound))
}

func TestIs404Error(t *testing.T) {
	assert.True(t, Is404Error(&scw.ResponseError{StatusCode: http.StatusNotFound}))
	assert.False(t, Is404Error(nil))
	assert.False(t, Is404Error(&scw.ResponseError{StatusCode: http.StatusBadRequest}))
}

func TestIs403Error(t *testing.T) {
	assert.True(t, Is403Error(&scw.ResponseError{StatusCode: http.StatusForbidden}))
	assert.False(t, Is403Error(nil))
	assert.False(t, Is403Error(&scw.ResponseError{StatusCode: http.StatusBadRequest}))
}
