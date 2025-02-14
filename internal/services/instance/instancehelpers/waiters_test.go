package instancehelpers

import (
	"testing"

	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/stretchr/testify/assert"
)

func TestUnknownVolume_VolumeTemplate(t *testing.T) {
	tests := []struct {
		name   string
		volume *UnknownVolume
		want   *instance.VolumeServerTemplate
	}{
		{
			name: "Root Volume",
			volume: &UnknownVolume{
				Name:               "",
				ID:                 "",
				InstanceVolumeType: instance.VolumeVolumeTypeLSSD,
				Size:               scw.SizePtr(20000000000),
				Boot:               scw.BoolPtr(false),
			},
			want: &instance.VolumeServerTemplate{
				Boot:       scw.BoolPtr(false),
				Size:       scw.SizePtr(20000000000),
				VolumeType: instance.VolumeVolumeTypeLSSD,
			},
		},
		{
			name: "Root Volume from ID",
			volume: &UnknownVolume{
				Name:               "tf-vol-stoic-johnson",
				ID:                 "25152794-d15a-4dd5-abfc-b19ec276aa20",
				InstanceVolumeType: instance.VolumeVolumeTypeLSSD,
				Size:               scw.SizePtr(20000000000),
				Boot:               scw.BoolPtr(true),
			},
			want: &instance.VolumeServerTemplate{
				ID:   scw.StringPtr("25152794-d15a-4dd5-abfc-b19ec276aa20"),
				Boot: scw.BoolPtr(true),
				Name: scw.StringPtr("tf-vol-stoic-johnson"),
			},
		},
		{
			name: "Additional Volume sbs",
			volume: &UnknownVolume{
				Name:               "tf-volume-elegant-minsky",
				ID:                 "cc380989-b71b-47f0-829f-062e329f4097",
				InstanceVolumeType: instance.VolumeVolumeTypeSbsVolume,
				Size:               scw.SizePtr(10000000000),
			},
			want: &instance.VolumeServerTemplate{
				ID:         scw.StringPtr("cc380989-b71b-47f0-829f-062e329f4097"),
				VolumeType: instance.VolumeVolumeTypeSbsVolume,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.volume.VolumeTemplate())
		})
	}
}
