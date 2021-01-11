package scaleway

import (
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/scaleway-sdk-go/scw"
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
			locality, id, err := parseLocalizedID(tc.localityID)
			if tc.err != "" {
				require.EqualError(t, err, tc.err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.locality, locality)
				assert.Equal(t, tc.id, id)
			}
		})
	}
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
			zone, id, err := parseZonedID(tc.localityID)
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
			region, id, err := parseRegionalID(tc.localityID)
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

func TestNewZonedId(t *testing.T) {
	assert.Equal(t, "fr-par-1/my-id", newZonedIDString(scw.ZoneFrPar1, "my-id"))
}

func TestNewRegionalId(t *testing.T) {
	assert.Equal(t, "fr-par/my-id", newRegionalIDString(scw.RegionFrPar, "my-id"))
}

func TestIsHTTPCodeError(t *testing.T) {
	assert.True(t, isHTTPCodeError(&scw.ResponseError{StatusCode: http.StatusBadRequest}, http.StatusBadRequest))
	assert.False(t, isHTTPCodeError(nil, http.StatusBadRequest))
	assert.False(t, isHTTPCodeError(&scw.ResponseError{StatusCode: http.StatusBadRequest}, http.StatusNotFound))
	assert.False(t, isHTTPCodeError(fmt.Errorf("not an http error"), http.StatusNotFound))
}

func TestIs404Error(t *testing.T) {
	assert.True(t, is404Error(&scw.ResponseError{StatusCode: http.StatusNotFound}))
	assert.False(t, is404Error(nil))
	assert.False(t, is404Error(&scw.ResponseError{StatusCode: http.StatusBadRequest}))
}

func TestIs403Error(t *testing.T) {
	assert.True(t, is403Error(&scw.ResponseError{StatusCode: http.StatusForbidden}))
	assert.False(t, is403Error(nil))
	assert.False(t, is403Error(&scw.ResponseError{StatusCode: http.StatusBadRequest}))
}

func TestGetRandomName(t *testing.T) {
	name := newRandomName("test")
	assert.True(t, strings.HasPrefix(name, "tf-test-"))
}

func testCheckResourceAttrFunc(name string, key string, test func(string) error) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource not found: %s", name)
		}
		value, ok := rs.Primary.Attributes[key]
		if !ok {
			return fmt.Errorf("key not found: %s", key)
		}
		err := test(value)
		if err != nil {
			return fmt.Errorf("test for %s %s did not pass test: %s", name, key, err)
		}
		return nil
	}
}

var UUIDRegex = regexp.MustCompile(`[0-9a-fA-F]{8}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{12}`)

func testCheckResourceAttrUUID(name string, key string) resource.TestCheckFunc {
	return resource.TestMatchResourceAttr(name, key, UUIDRegex)
}

func testCheckResourceAttrIPv4(name string, key string) resource.TestCheckFunc {
	return testCheckResourceAttrFunc(name, key, func(value string) error {
		ip := net.ParseIP(value)
		if ip.To4() == nil {
			return fmt.Errorf("%s is not a valid IPv4", value)
		}
		return nil
	})
}

func testCheckResourceAttrIPv6(name string, key string) resource.TestCheckFunc {
	return testCheckResourceAttrFunc(name, key, func(value string) error {
		ip := net.ParseIP(value)
		if ip.To16() == nil {
			return fmt.Errorf("%s is not a valid IPv6", value)
		}
		return nil
	})
}
