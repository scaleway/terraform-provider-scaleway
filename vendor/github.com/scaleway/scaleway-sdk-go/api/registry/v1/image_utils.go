package registry

import (
	"time"

	"github.com/scaleway/scaleway-sdk-go/internal/async"
	"github.com/scaleway/scaleway-sdk-go/internal/errors"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// WaitForNamespaceRequest is used by WaitForNamespace method
type WaitForImageRequest struct {
	ImageID       string
	Region        scw.Region
	Timeout       *time.Duration
	RetryInterval *time.Duration
}

// WaitForImage wait for the image to be in a "terminal state" before returning.
// This function can be used to wait for an image to be ready for example.
func (s *API) WaitForImage(req *WaitForImageRequest) (*Image, error) {
	timeout := defaultTimeout
	if req.Timeout != nil {
		timeout = *req.Timeout
	}
	retryInterval := defaultRetryInterval
	if req.RetryInterval != nil {
		retryInterval = *req.RetryInterval
	}

	terminalStatus := map[ImageStatus]struct{}{
		ImageStatusReady:   {},
		ImageStatusLocked:  {},
		ImageStatusError:   {},
		ImageStatusUnknown: {},
	}

	image, err := async.WaitSync(&async.WaitSyncConfig{
		Get: func() (interface{}, bool, error) {
			img, err := s.GetImage(&GetImageRequest{
				Region:  req.Region,
				ImageID: req.ImageID,
			})
			if err != nil {
				return nil, false, err
			}

			_, isTerminal := terminalStatus[img.Status]

			return img, isTerminal, err
		},
		Timeout:          timeout,
		IntervalStrategy: async.LinearIntervalStrategy(retryInterval),
	})
	if err != nil {
		return nil, errors.Wrap(err, "waiting for image failed")
	}
	return image.(*Image), nil
}
