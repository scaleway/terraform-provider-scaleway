package instance_test

import (
	"testing"

	instanceSDK "github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/instance"
	"github.com/stretchr/testify/assert"
)

func TestUnknownVolume_VolumeTemplate(t *testing.T) {
	type fields struct {
		Zone               scw.Zone
		ID                 string
		Name               string
		Size               *scw.Size
		ServerID           *string
		Boot               *bool
		IsBlockVolume      bool
		InstanceVolumeType instanceSDK.VolumeVolumeType
	}
	tests := []struct {
		name   string
		volume *instance.UnknownVolume
		want   *instanceSDK.VolumeServerTemplate
	}{
		{
			name: "Root Volume",
			volume: &instance.UnknownVolume{
				Name:               "",
				ID:                 "",
				InstanceVolumeType: instanceSDK.VolumeVolumeTypeLSSD,
				Size:               scw.SizePtr(20000000000),
				Boot:               scw.BoolPtr(false),
			},
			want: &instanceSDK.VolumeServerTemplate{
				Boot:       scw.BoolPtr(false),
				Size:       scw.SizePtr(20000000000),
				VolumeType: instanceSDK.VolumeVolumeTypeLSSD,
			},
		},
		{
			name: "Root Volume from ID",
			volume: &instance.UnknownVolume{
				Name:               "tf-vol-stoic-johnson",
				ID:                 "25152794-d15a-4dd5-abfc-b19ec276aa20",
				InstanceVolumeType: instanceSDK.VolumeVolumeTypeLSSD,
				Size:               scw.SizePtr(20000000000),
				Boot:               scw.BoolPtr(true),
			},
			want: &instanceSDK.VolumeServerTemplate{
				ID:   scw.StringPtr("25152794-d15a-4dd5-abfc-b19ec276aa20"),
				Boot: scw.BoolPtr(true),
				Name: scw.StringPtr("tf-vol-stoic-johnson"),
			},
		},
		{
			name: "Additional Volume sbs",
			volume: &instance.UnknownVolume{
				Name:               "tf-volume-elegant-minsky",
				ID:                 "cc380989-b71b-47f0-829f-062e329f4097",
				InstanceVolumeType: instanceSDK.VolumeVolumeTypeSbsVolume,
				Size:               scw.SizePtr(10000000000),
				IsBlockVolume:      true,
			},
			want: &instanceSDK.VolumeServerTemplate{
				ID:         scw.StringPtr("cc380989-b71b-47f0-829f-062e329f4097"),
				VolumeType: instanceSDK.VolumeVolumeTypeSbsVolume,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.volume.VolumeTemplate())
		})
	}
}
