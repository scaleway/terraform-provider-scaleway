package k8s

import (
	"time"

	"github.com/scaleway/scaleway-sdk-go/internal/async"
	"github.com/scaleway/scaleway-sdk-go/internal/errors"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	waitForClusterDefaultTimeout = time.Minute * 15
	waitForPoolDefaultTimeout    = time.Minute * 15
)

// WaitForClusterRequest is used by WaitForCluster method.
type WaitForClusterRequest struct {
	ClusterID string
	Region    scw.Region
	Status    ClusterStatus
	Timeout   *time.Duration
}

// WaitForCluster waits for the cluster to be in a "terminal state" before returning.
func (s *API) WaitForCluster(req *WaitForClusterRequest) (*Cluster, error) {
	timeout := waitForClusterDefaultTimeout
	if req.Timeout != nil {
		timeout = *req.Timeout
	}
	terminalStatus := map[ClusterStatus]struct{}{
		ClusterStatusReady:   {},
		ClusterStatusError:   {},
		ClusterStatusWarning: {},
		ClusterStatusLocked:  {},
		ClusterStatusDeleted: {},
	}

	cluster, err := async.WaitSync(&async.WaitSyncConfig{
		Get: func() (interface{}, bool, error) {
			cluster, err := s.GetCluster(&GetClusterRequest{
				ClusterID: req.ClusterID,
				Region:    req.Region,
			})
			if err != nil {
				return nil, false, err
			}

			_, isTerminal := terminalStatus[cluster.Status]
			return cluster, isTerminal, nil
		},
		Timeout:          timeout,
		IntervalStrategy: async.LinearIntervalStrategy(5 * time.Second),
	})
	if err != nil {
		return nil, errors.Wrap(err, "waiting for cluster failed")
	}
	return cluster.(*Cluster), nil
}

// WaitForPoolRequest is used by WaitForPool method.
type WaitForPoolRequest struct {
	PoolID  string
	Region  scw.Region
	Timeout *time.Duration
}

// WaitForPool waits for a pool to be ready
func (s *API) WaitForPool(req *WaitForPoolRequest) (*Pool, error) {
	terminalStatus := map[PoolStatus]struct{}{
		PoolStatusReady:   {},
		PoolStatusWarning: {},
	}

	timeout := waitForPoolDefaultTimeout
	if req.Timeout != nil {
		timeout = *req.Timeout
	}

	pool, err := async.WaitSync(&async.WaitSyncConfig{
		Get: func() (interface{}, bool, error) {
			res, err := s.GetPool(&GetPoolRequest{
				PoolID: req.PoolID,
				Region: req.Region,
			})

			if err != nil {
				return nil, false, err
			}
			_, isTerminal := terminalStatus[res.Status]

			return res, isTerminal, nil
		},
		Timeout:          timeout,
		IntervalStrategy: async.LinearIntervalStrategy(5 * time.Second),
	})

	return pool.(*Pool), err
}
