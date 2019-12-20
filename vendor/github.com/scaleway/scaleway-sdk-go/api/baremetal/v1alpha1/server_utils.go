package baremetal

import (
	"time"

	"github.com/scaleway/scaleway-sdk-go/internal/async"
	"github.com/scaleway/scaleway-sdk-go/internal/errors"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// WaitForServerRequest is used by WaitForServer method.
type WaitForServerRequest struct {
	ServerID string
	Zone     scw.Zone
	Timeout  time.Duration
}

// WaitForServer wait for the server to be in a "terminal state" before returning.
// This function can be used to wait for a server to be created.
func (s *API) WaitForServer(req *WaitForServerRequest) (*Server, error) {
	terminalStatus := map[ServerStatus]struct{}{
		ServerStatusReady:   {},
		ServerStatusStopped: {},
		ServerStatusError:   {},
		ServerStatusLocked:  {},
		ServerStatusUnknown: {},
	}

	server, err := async.WaitSync(&async.WaitSyncConfig{
		Get: func() (interface{}, error, bool) {
			res, err := s.GetServer(&GetServerRequest{
				ServerID: req.ServerID,
				Zone:     req.Zone,
			})
			if err != nil {
				return nil, err, false
			}

			_, isTerminal := terminalStatus[res.Status]
			return res, err, isTerminal
		},
		Timeout:          req.Timeout,
		IntervalStrategy: async.LinearIntervalStrategy(5 * time.Second),
	})
	if err != nil {
		return nil, errors.Wrap(err, "waiting for server failed")
	}

	return server.(*Server), nil
}

// WaitForServerInstallRequest is used by WaitForServerInstall method.
type WaitForServerInstallRequest struct {
	ServerID string
	Zone     scw.Zone
	Timeout  time.Duration
}

// WaitForServerInstall wait for the server install to be in a
// "terminal state" before returning.
// This function can be used to wait for a server to be installed.
func (s *API) WaitForServerInstall(req *WaitForServerInstallRequest) (*Server, error) {
	installTerminalStatus := map[ServerInstallStatus]struct{}{
		ServerInstallStatusCompleted: {},
		ServerInstallStatusError:     {},
		ServerInstallStatusUnknown:   {},
	}

	server, err := async.WaitSync(&async.WaitSyncConfig{
		Get: func() (interface{}, error, bool) {
			res, err := s.GetServer(&GetServerRequest{
				ServerID: req.ServerID,
				Zone:     req.Zone,
			})
			if err != nil {
				return nil, err, false
			}

			if res.Install == nil {
				return nil, errors.New("server creation has not begun for server %s", req.ServerID), false
			}

			_, isTerminal := installTerminalStatus[res.Install.Status]
			return res, err, isTerminal
		},
		Timeout:          req.Timeout,
		IntervalStrategy: async.LinearIntervalStrategy(15 * time.Second),
	})
	if err != nil {
		return nil, errors.Wrap(err, "waiting for server installation failed")
	}

	return server.(*Server), nil
}
