package instance

import (
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/api/marketplace/v2"
)

func volumeTypeToMarketplaceFilter(volumeType instance.VolumeVolumeType) marketplace.LocalImageType {
	switch volumeType {
	case instance.VolumeVolumeTypeSbsVolume:
		return marketplace.LocalImageTypeInstanceSbs
	case instance.VolumeVolumeTypeBSSD:
		return marketplace.LocalImageTypeInstanceLocal
	case instance.VolumeVolumeTypeLSSD:
		return marketplace.LocalImageTypeInstanceLocal
	}

	return marketplace.LocalImageTypeInstanceSbs
}
