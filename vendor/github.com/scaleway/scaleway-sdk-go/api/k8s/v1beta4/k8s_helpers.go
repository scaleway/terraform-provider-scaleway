package k8s

import (
	"time"

	"github.com/scaleway/scaleway-sdk-go/internal/async"
	"github.com/scaleway/scaleway-sdk-go/internal/errors"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// WaitForClusterRequest is used by WaitForCluster method.
type WaitForClusterRequest struct {
	ClusterID string
	Region    scw.Region
	Status    ClusterStatus
	Timeout   time.Duration
}

// WaitForCluster waits for the cluster to be in a "terminal state" before returning.
func (s *API) WaitForCluster(req *WaitForClusterRequest) (*Cluster, error) {
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
		Timeout:          req.Timeout,
		IntervalStrategy: async.LinearIntervalStrategy(5 * time.Second),
	})
	if err != nil {
		return nil, errors.Wrap(err, "waiting for cluster failed")
	}
	return cluster.(*Cluster), nil
}
