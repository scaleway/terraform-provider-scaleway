package instance

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	instanceSDK "github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetServerVolumesState_NoVolumes(t *testing.T) {
	d := schema.TestResourceDataRaw(t, serverSchema(), map[string]any{
		"type":  "DEV1-S",
		"image": "ubuntu_focal",
	})
	volumes := map[string]*instanceSDK.VolumeServer{}

	err := setServerVolumesState(d, nil, scw.ZoneFrPar1, volumes)
	require.NoError(t, err)

	assert.Empty(t, d.Get("root_volume"))
	assert.Empty(t, d.Get("additional_volume_ids"))
}

func TestSetServerVolumesState_NilVolumes(t *testing.T) {
	d := schema.TestResourceDataRaw(t, serverSchema(), map[string]any{
		"type":  "DEV1-S",
		"image": "ubuntu_focal",
	})

	err := setServerVolumesState(d, nil, scw.ZoneFrPar1, nil)
	require.NoError(t, err)

	assert.Empty(t, d.Get("root_volume"))
	assert.Empty(t, d.Get("additional_volume_ids"))
}

func TestSetServerVolumesState_OnlyAdditionalVolume(t *testing.T) {
	d := schema.TestResourceDataRaw(t, serverSchema(), map[string]any{
		"type":  "DEV1-S",
		"image": "ubuntu_focal",
	})
	volumes := map[string]*instanceSDK.VolumeServer{
		"1": {ID: "11111111-1111-1111-1111-111111111111"},
	}

	err := setServerVolumesState(d, nil, scw.ZoneFrPar1, volumes)
	require.NoError(t, err)

	assert.Empty(t, d.Get("root_volume"))
	additional := d.Get("additional_volume_ids").([]any)
	require.Len(t, additional, 1)
	assert.Equal(t, "fr-par-1/11111111-1111-1111-1111-111111111111", additional[0])
}

func TestSortVolumeServer_EmptyAndNil(t *testing.T) {
	assert.Empty(t, sortVolumeServer(nil))
	assert.Empty(t, sortVolumeServer(map[string]*instanceSDK.VolumeServer{}))
}

func TestSortVolumeServer_Order(t *testing.T) {
	v0 := &instanceSDK.VolumeServer{ID: "vol-0"}
	v1 := &instanceSDK.VolumeServer{ID: "vol-1"}
	v2 := &instanceSDK.VolumeServer{ID: "vol-2"}

	volumes := map[string]*instanceSDK.VolumeServer{
		"2": v2,
		"0": v0,
		"1": v1,
	}

	sorted := sortVolumeServer(volumes)
	require.Len(t, sorted, 3)
	assert.Equal(t, "vol-0", sorted[0].ID)
	assert.Equal(t, "vol-1", sorted[1].ID)
	assert.Equal(t, "vol-2", sorted[2].ID)
}
