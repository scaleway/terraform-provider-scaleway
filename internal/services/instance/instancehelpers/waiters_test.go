package instancehelpers_test

import (
	"testing"

	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/instance/instancehelpers"
	"github.com/stretchr/testify/assert"
)

func TestUnknownVolume_VolumeTemplate(t *testing.T) {
	tests := []struct {
		volume *instancehelpers.UnknownVolume
		want   *instance.VolumeServerTemplate
		name   string
	}{
		{
			name: "Root Volume",
			volume: &instancehelpers.UnknownVolume{
				Name:               "",
				ID:                 "",
				InstanceVolumeType: instance.VolumeVolumeTypeLSSD,
				Size:               scw.SizePtr(20000000000),
				Boot:               new(false),
			},
			want: &instance.VolumeServerTemplate{
				Boot:       new(false),
				Size:       scw.SizePtr(20000000000),
				VolumeType: instance.VolumeVolumeTypeLSSD,
			},
		},
		{
			name: "Root Volume from ID",
			volume: &instancehelpers.UnknownVolume{
				Name:               "tf-vol-stoic-johnson",
				ID:                 "25152794-d15a-4dd5-abfc-b19ec276aa20",
				InstanceVolumeType: instance.VolumeVolumeTypeLSSD,
				Size:               scw.SizePtr(20000000000),
				Boot:               new(true),
			},
			want: &instance.VolumeServerTemplate{
				ID:   new("25152794-d15a-4dd5-abfc-b19ec276aa20"),
				Boot: new(true),
				Name: new("tf-vol-stoic-johnson"),
			},
		},
		{
			name: "Additional Volume sbs",
			volume: &instancehelpers.UnknownVolume{
				Name:               "tf-volume-elegant-minsky",
				ID:                 "cc380989-b71b-47f0-829f-062e329f4097",
				InstanceVolumeType: instance.VolumeVolumeTypeSbsVolume,
				Size:               scw.SizePtr(10000000000),
			},
			want: &instance.VolumeServerTemplate{
				ID:         new("cc380989-b71b-47f0-829f-062e329f4097"),
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
