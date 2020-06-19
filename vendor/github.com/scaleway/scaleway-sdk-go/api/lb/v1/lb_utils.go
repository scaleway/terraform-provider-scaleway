package lb

import (
	"time"

	"github.com/scaleway/scaleway-sdk-go/internal/async"
	"github.com/scaleway/scaleway-sdk-go/internal/errors"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	defaultRetryInterval = 2 * time.Second
	defaultTimeout       = 5 * time.Minute
)

// WaitForLbRequest is used by WaitForLb method.
type WaitForLbRequest struct {
	LbID          string
	Region        scw.Region
	Timeout       *time.Duration
	RetryInterval *time.Duration
}

// WaitForLb waits for the lb to be in a "terminal state" before returning.
// This function can be used to wait for a lb to be ready for example.
func (s *API) WaitForLb(req *WaitForLbRequest) (*Lb, error) {
	timeout := defaultTimeout
	if req.Timeout != nil {
		timeout = *req.Timeout
	}
	retryInterval := defaultRetryInterval
	if req.RetryInterval != nil {
		retryInterval = *req.RetryInterval
	}

	terminalStatus := map[LbStatus]struct{}{
		LbStatusReady:   {},
		LbStatusStopped: {},
		LbStatusError:   {},
		LbStatusLocked:  {},
	}

	lb, err := async.WaitSync(&async.WaitSyncConfig{
		Get: func() (interface{}, bool, error) {
			res, err := s.GetLb(&GetLbRequest{
				LbID:   req.LbID,
				Region: req.Region,
			})

			if err != nil {
				return nil, false, err
			}
			_, isTerminal := terminalStatus[res.Status]

			return res, isTerminal, nil
		},
		Timeout:          timeout,
		IntervalStrategy: async.LinearIntervalStrategy(retryInterval),
	})
	if err != nil {
		return nil, errors.Wrap(err, "waiting for lb failed")
	}
	return lb.(*Lb), nil
}
