package k8s

import (
	"time"

	"github.com/scaleway/scaleway-sdk-go/internal/async"
	"github.com/scaleway/scaleway-sdk-go/internal/errors"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	waitForClusterDefaultTimeout = 15 * time.Minute
	waitForPoolDefaultTimeout    = 15 * time.Minute
	waitForNodeDefaultTimeout    = 15 * time.Minute
	defaultRetryInterval         = 5 * time.Second
)

// WaitForClusterRequest is used by WaitForCluster method.
type WaitForClusterRequest struct {
	ClusterID     string
	Region        scw.Region
	Status        ClusterStatus
	Timeout       *time.Duration
	RetryInterval *time.Duration
}

// WaitForCluster waits for the cluster to be in a "terminal state" before returning.
func (s *API) WaitForCluster(req *WaitForClusterRequest) (*Cluster, error) {
	timeout := *req.Timeout
	if timeout == 0 {
		timeout = waitForClusterDefaultTimeout
	}
	retryInterval := defaultRetryInterval
	if req.RetryInterval != nil {
		retryInterval = *req.RetryInterval
	}

	terminalStatus := map[ClusterStatus]struct{}{
		ClusterStatusReady:        {},
		ClusterStatusLocked:       {},
		ClusterStatusDeleted:      {},
		ClusterStatusPoolRequired: {},
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
		IntervalStrategy: async.LinearIntervalStrategy(retryInterval),
	})
	if err != nil {
		return nil, errors.Wrap(err, "waiting for cluster failed")
	}

	return cluster.(*Cluster), nil
}

// WaitForPoolRequest is used by WaitForPool method.
type WaitForPoolRequest struct {
	PoolID        string
	Region        scw.Region
	Timeout       *time.Duration
	RetryInterval *time.Duration
}

// WaitForPool waits for a pool to be ready
func (s *API) WaitForPool(req *WaitForPoolRequest) (*Pool, error) {
	timeout := *req.Timeout
	if timeout == 0 {
		timeout = waitForPoolDefaultTimeout
	}
	retryInterval := defaultRetryInterval
	if req.RetryInterval != nil {
		retryInterval = *req.RetryInterval
	}

	terminalStatus := map[PoolStatus]struct{}{
		PoolStatusReady:   {},
		PoolStatusWarning: {},
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
		IntervalStrategy: async.LinearIntervalStrategy(retryInterval),
	})
	if err != nil {
		return nil, errors.Wrap(err, "waiting for pool failed")
	}

	return pool.(*Pool), err
}

// WaitForNodeRequest is used by WaitForNode method.
type WaitForNodeRequest struct {
	NodeID        string
	Region        scw.Region
	Timeout       *time.Duration
	RetryInterval *time.Duration
}

// WaitForNode waits for a Node to be ready
func (s *API) WaitForNode(req *WaitForNodeRequest) (*Node, error) {
	timeout := waitForNodeDefaultTimeout
	if req.Timeout != nil {
		timeout = *req.Timeout
	}
	retryInterval := defaultRetryInterval
	if req.RetryInterval != nil {
		retryInterval = *req.RetryInterval
	}

	terminalStatus := map[NodeStatus]struct{}{
		NodeStatusCreationError: {},
		NodeStatusReady:         {},
	}

	node, err := async.WaitSync(&async.WaitSyncConfig{
		Get: func() (interface{}, bool, error) {
			res, err := s.GetNode(&GetNodeRequest{
				NodeID: req.NodeID,
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
		return nil, errors.Wrap(err, "waiting for node failed")
	}

	return node.(*Node), err
}
