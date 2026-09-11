//nolint:testpackage // Tests need access to unexported messageq flatten/expand helpers.
package messageq

import (
	"testing"

	messageqapi "github.com/scaleway/scaleway-sdk-go/api/messageq/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVolumeSizeConversion(t *testing.T) {
	t.Parallel()

	// The MessageQ API expresses sizes in decimal bytes: 5 GB is 5e9 bytes, not 5 GiB.
	assert.Equal(t, scw.Size(5_000_000_000), expandVolumeSizeBytes(5))
	assert.Equal(t, 5, bytesToGB(5_000_000_000))
	assert.Equal(t, 16, bytesToGB(16_000_000_000))
	assert.Equal(t, 0, bytesToGB(0))
}

func TestRegionAndIDFromAttr(t *testing.T) {
	t.Parallel()

	const uuid = "11111111-1111-1111-1111-111111111111"

	t.Run("localized id overrides the fallback region", func(t *testing.T) {
		t.Parallel()

		region, id := regionAndIDFromAttr("nl-ams/"+uuid, scw.RegionFrPar)
		assert.Equal(t, scw.RegionNlAms, region)
		assert.Equal(t, uuid, id)
	})

	t.Run("bare uuid keeps the fallback region", func(t *testing.T) {
		t.Parallel()

		region, id := regionAndIDFromAttr(uuid, scw.RegionFrPar)
		assert.Equal(t, scw.RegionFrPar, region)
		assert.Equal(t, uuid, id)
	})
}

func TestExpandEndpointSpecsFromPrivateNetwork(t *testing.T) {
	t.Parallel()

	t.Run("no private network requests a public endpoint", func(t *testing.T) {
		t.Parallel()

		specs := expandEndpointSpecsFromPrivateNetwork("")
		require.Len(t, specs, 1)
		require.NotNil(t, specs[0].Public)
		assert.Nil(t, specs[0].PrivateNetwork)
	})

	t.Run("private network id requests a private endpoint", func(t *testing.T) {
		t.Parallel()

		specs := expandEndpointSpecsFromPrivateNetwork("pn-id")
		require.Len(t, specs, 1)
		require.NotNil(t, specs[0].PrivateNetwork)
		assert.Equal(t, "pn-id", specs[0].PrivateNetwork.PrivateNetworkID)
		assert.Nil(t, specs[0].Public)
	})
}

func TestFlattenEndpoints(t *testing.T) {
	t.Parallel()

	assert.Nil(t, flattenEndpoints(nil))

	endpoints := []*messageqapi.Endpoint{
		{
			ID:     "public-endpoint",
			Public: &messageqapi.EndpointPublicDetails{},
			Services: []*messageqapi.EndpointService{
				{Name: "amqp", Port: 5672, URL: "amqps://example.com:5672"},
			},
		},
		{
			ID:             "private-endpoint",
			PrivateNetwork: &messageqapi.EndpointPrivateNetworkDetails{PrivateNetworkID: "pn-id"},
		},
	}

	flattened := flattenEndpoints(endpoints)
	require.Len(t, flattened, 2)

	assert.Equal(t, "public-endpoint", flattened[0]["id"])
	assert.Equal(t, true, flattened[0]["public"])
	assert.Equal(t, []map[string]any{
		{"name": "amqp", "port": 5672, "url": "amqps://example.com:5672"},
	}, flattened[0]["services"])

	assert.Equal(t, "private-endpoint", flattened[1]["id"])
	assert.Equal(t, false, flattened[1]["public"])
	assert.Equal(t, "pn-id", flattened[1]["private_network_id"])
	// An endpoint without services must not report an empty list.
	assert.NotContains(t, flattened[1], "services")
}

func TestResourceUserParseID(t *testing.T) {
	t.Parallel()

	region, deploymentID, userName, err := ResourceUserParseID("fr-par/deployment-id/my-user")
	require.NoError(t, err)
	assert.Equal(t, scw.RegionFrPar, region)
	assert.Equal(t, "deployment-id", deploymentID)
	assert.Equal(t, "my-user", userName)

	// A malformed ID must surface an error instead of silently querying an empty region.
	for _, malformed := range []string{"", "fr-par", "fr-par/deployment-id"} {
		_, _, _, err := ResourceUserParseID(malformed)
		assert.Error(t, err, "expected an error for %q", malformed)
	}
}
