package messageq_test

import (
	"testing"

	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/messageq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVolumeSizeConversion(t *testing.T) {
	t.Parallel()

	// The MessageQ API expresses sizes in decimal bytes: 5 GB is 5e9 bytes, not 5 GiB.
	assert.Equal(t, scw.Size(5_000_000_000), messageq.ExpandVolumeSizeBytes(5))
	assert.Equal(t, 5, messageq.BytesToGB(5_000_000_000))
	assert.Equal(t, 16, messageq.BytesToGB(16_000_000_000))
	assert.Equal(t, 0, messageq.BytesToGB(0))
}

func TestRegionAndIDFromAttr(t *testing.T) {
	t.Parallel()

	const uuid = "11111111-1111-1111-1111-111111111111"

	t.Run("localized id overrides the fallback region", func(t *testing.T) {
		t.Parallel()

		region, id, err := messageq.RegionAndIDFromAttr("nl-ams/"+uuid, scw.RegionFrPar)
		require.NoError(t, err)
		assert.Equal(t, scw.RegionNlAms, region)
		assert.Equal(t, uuid, id)
	})

	t.Run("bare uuid keeps the fallback region", func(t *testing.T) {
		t.Parallel()

		region, id, err := messageq.RegionAndIDFromAttr(uuid, scw.RegionFrPar)
		require.NoError(t, err)
		assert.Equal(t, scw.RegionFrPar, region)
		assert.Equal(t, uuid, id)
	})

	t.Run("malformed localized id returns an error", func(t *testing.T) {
		t.Parallel()

		_, _, err := messageq.RegionAndIDFromAttr("not-a-region/"+uuid, scw.RegionFrPar)
		assert.Error(t, err)
	})
}

func TestExpandEndpointSpecsFromPrivateNetwork(t *testing.T) {
	t.Parallel()

	t.Run("no private network requests a public endpoint", func(t *testing.T) {
		t.Parallel()

		specs := messageq.ExpandEndpointSpecsFromPrivateNetwork("")
		require.Len(t, specs, 1)
		require.NotNil(t, specs[0].Public)
		assert.Nil(t, specs[0].PrivateNetwork)
	})

	t.Run("private network id requests a private endpoint", func(t *testing.T) {
		t.Parallel()

		specs := messageq.ExpandEndpointSpecsFromPrivateNetwork("pn-id")
		require.Len(t, specs, 1)
		require.NotNil(t, specs[0].PrivateNetwork)
		assert.Equal(t, "pn-id", specs[0].PrivateNetwork.PrivateNetworkID)
		assert.Nil(t, specs[0].Public)
	})
}

func TestResourceUserParseID(t *testing.T) {
	t.Parallel()

	region, deploymentID, userName, err := messageq.ResourceUserParseID("fr-par/deployment-id/my-user")
	require.NoError(t, err)
	assert.Equal(t, scw.RegionFrPar, region)
	assert.Equal(t, "deployment-id", deploymentID)
	assert.Equal(t, "my-user", userName)

	// A malformed ID must surface an error instead of silently querying an empty region.
	for _, malformed := range []string{"", "fr-par", "fr-par/deployment-id"} {
		_, _, _, err := messageq.ResourceUserParseID(malformed)
		assert.Error(t, err, "expected an error for %q", malformed)
	}
}
