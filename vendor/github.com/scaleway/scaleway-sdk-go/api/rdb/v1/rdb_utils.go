package rdb

import (
	"time"

	"github.com/scaleway/scaleway-sdk-go/internal/async"
	"github.com/scaleway/scaleway-sdk-go/internal/errors"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// WaitForInstanceRequest is used by WaitForInstance method.
type WaitForInstanceRequest struct {
	InstanceID string
	Region     scw.Region
	Timeout    time.Duration
}

// WaitForInstance waits for the instance to be in a "terminal state" before returning.
// This function can be used to wait for an instance to be ready for example.
func (s *API) WaitForInstance(req *WaitForInstanceRequest) (*Instance, error) {
	terminalStatus := map[InstanceStatus]struct{}{
		InstanceStatusReady:    {},
		InstanceStatusDiskFull: {},
		InstanceStatusError:    {},
	}

	instance, err := async.WaitSync(&async.WaitSyncConfig{
		Get: func() (interface{}, bool, error) {
			res, err := s.GetInstance(&GetInstanceRequest{
				InstanceID: req.InstanceID,
				Region:     req.Region,
			})

			if err != nil {
				return nil, false, err
			}
			_, isTerminal := terminalStatus[res.Status]

			return res, isTerminal, nil
		},
		Timeout:          req.Timeout,
		IntervalStrategy: async.LinearIntervalStrategy(5 * time.Second),
	})
	if err != nil {
		return nil, errors.Wrap(err, "waiting for instance failed")
	}
	return instance.(*Instance), nil
}
