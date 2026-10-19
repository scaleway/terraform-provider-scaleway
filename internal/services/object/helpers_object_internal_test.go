package object

import (
	"testing"

	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRetrieveS3EndpointProfileFromMeta tests that retrieveS3Endpoint
// resolves the S3 endpoint from the scw profile (via the scw client) when
// called through NewS3ClientFromMeta's code path (m == nil, metaStruct != nil).
func TestRetrieveS3EndpointProfileFromMeta(t *testing.T) {
	for _, env := range []string{
		scw.ScwS3EndpointEnv,
		scw.ScwAccessKeyEnv,
		scw.ScwSecretKeyEnv,
		scw.ScwDefaultProjectIDEnv,
		scw.ScwDefaultOrganizationIDEnv,
		scw.ScwDefaultRegionEnv,
		scw.ScwDefaultZoneEnv,
		"SCW_CONFIG_PATH",
	} {
		t.Setenv(env, "")
	}

	profileEndpoint := "https://profile-s3.example.com"

	profile := &scw.Profile{
		AccessKey:     new("SCWXXXXXXXXXXXXXXXXX"),
		SecretKey:     new("866f4a9a-d058-4d3c-a39f-86930849ccc0"),
		DefaultRegion: new(scw.RegionFrPar.String()),
		DefaultZone:   new(scw.ZoneFrPar1.String()),
		S3Endpoint:    new(profileEndpoint),
	}

	m, err := meta.NewMetaFromProfile(t.Context(), profile, nil, nil, nil, "test", nil)
	require.NoError(t, err)

	// Simulate the NewS3ClientFromMeta call path: metaStruct=m, d=nil, m=nil
	endpoint, err := retrieveS3Endpoint(t.Context(), m, nil, nil, scw.RegionFrPar.String())
	require.NoError(t, err)

	assert.Equal(t, profileEndpoint, endpoint)
}
