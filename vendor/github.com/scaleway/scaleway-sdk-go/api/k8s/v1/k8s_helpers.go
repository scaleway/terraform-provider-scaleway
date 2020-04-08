package k8s

import (
	"time"

	"github.com/scaleway/scaleway-sdk-go/internal/async"
	"github.com/scaleway/scaleway-sdk-go/internal/errors"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

var (
	// RetryInterval is needed when running recorded tests (e.g. on scaleway-cli)
	// it allows to execute the WaitFor funcs immediately
	RetryInterval = defaultRetryInterval
)

const (
	waitForClusterDefaultTimeout = time.Minute * 15
	waitForPoolDefaultTimeout    = time.Minute * 15
	waitForNodeDefaultTimeout    = time.Minute * 15
	defaultRetryInterval         = time.Second * 5
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
		IntervalStrategy: async.LinearIntervalStrategy(RetryInterval),
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
		IntervalStrategy: async.LinearIntervalStrategy(RetryInterval),
	})
	if err != nil {
		return nil, errors.Wrap(err, "waiting for pool failed")
	}

	return pool.(*Pool), err
}

// WaitForNodeRequest is used by WaitForNode method.
type WaitForNodeRequest struct {
	NodeID  string
	Region  scw.Region
	Timeout *time.Duration
}

// WaitForNode waits for a Node to be ready
func (s *API) WaitForNode(req *WaitForNodeRequest) (*Node, error) {
	terminalStatus := map[NodeStatus]struct{}{
		NodeStatusCreationError: {},
		NodeStatusReady:         {},
	}

	timeout := waitForNodeDefaultTimeout
	if req.Timeout != nil {
		timeout = *req.Timeout
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
		IntervalStrategy: async.LinearIntervalStrategy(RetryInterval),
	})
	if err != nil {
		return nil, errors.Wrap(err, "waiting for node failed")
	}

	return node.(*Node), err
}
