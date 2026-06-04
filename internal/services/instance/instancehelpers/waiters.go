package instancehelpers

import (
	"context"
	"time"

	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

const DefaultInstanceRetryInterval = 5 * time.Second

func WaitForVolume(ctx context.Context, api *instance.API, zone scw.Zone, id string, timeout time.Duration) (*instance.Volume, error) {
	retryInterval := DefaultInstanceRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	volume, err := api.WaitForVolume(&instance.WaitForVolumeRequest{
		VolumeID:      id,
		Zone:          zone,
		Timeout:       new(timeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))

	return volume, err
}
