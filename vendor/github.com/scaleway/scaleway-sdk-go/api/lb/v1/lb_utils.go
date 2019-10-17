package lb

import (
	"time"

	"github.com/scaleway/scaleway-sdk-go/internal/async"
	"github.com/scaleway/scaleway-sdk-go/internal/errors"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// WaitForLbRequest is used by WaitForLb method.
type WaitForLbRequest struct {
	LbID    string
	Region  scw.Region
	Timeout time.Duration
}

// WaitForLb waits for the lb to be in a "terminal state" before returning.
// This function can be used to wait for a lb to be ready for example.
func (s *API) WaitForLb(req *WaitForLbRequest) (*Lb, error) {

	terminalStatus := map[LbStatus]struct{}{
		LbStatusReady:   {},
		LbStatusStopped: {},
		LbStatusError:   {},
		LbStatusLocked:  {},
	}

	lb, err := async.WaitSync(&async.WaitSyncConfig{
		Get: func() (interface{}, error, bool) {
			res, err := s.GetLb(&GetLbRequest{
				LbID:   req.LbID,
				Region: req.Region,
			})

			if err != nil {
				return nil, err, false
			}
			_, isTerminal := terminalStatus[res.Status]

			return res, nil, isTerminal
		},
		Timeout:          req.Timeout,
		IntervalStrategy: async.LinearIntervalStrategy(5 * time.Second),
	})
	if err != nil {
		return nil, errors.Wrap(err, "waiting for lb failed")
	}
	return lb.(*Lb), nil
}
