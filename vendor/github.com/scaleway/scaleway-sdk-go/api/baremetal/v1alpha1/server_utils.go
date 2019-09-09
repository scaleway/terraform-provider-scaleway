package baremetal

import (
	"time"

	"github.com/scaleway/scaleway-sdk-go/internal/async"
	"github.com/scaleway/scaleway-sdk-go/internal/errors"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// WaitForServerRequest is used by WaitForServer method
type WaitForServerRequest struct {
	ServerID string
	Zone     scw.Zone
	Timeout  time.Duration
}

// WaitForServer wait for the server to be in a "terminal state" before returning.
// This function can be used to wait for a server to be installed for example.
func (s *API) WaitForServer(req *WaitForServerRequest) (*Server, scw.SdkError) {

	terminalStatus := map[ServerStatus]struct{}{
		ServerStatusReady:   {},
		ServerStatusStopped: {},
		ServerStatusError:   {},
		ServerStatusUnknown: {},
	}

	installTerminalStatus := map[ServerInstallStatus]struct{}{
		ServerInstallStatusCompleted: {},
		ServerInstallStatusToInstall: {},
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
			_, isTerminal := terminalStatus[res.Status]
			isTerminalInstall := true
			if res.Install != nil {
				_, isTerminalInstall = installTerminalStatus[res.Install.Status]
			}

			return res, err, isTerminal && isTerminalInstall
		},
		Timeout:          req.Timeout,
		IntervalStrategy: async.LinearIntervalStrategy(5 * time.Second),
	})
	if err != nil {
		return nil, errors.Wrap(err, "waiting for server failed")
	}

	return server.(*Server), nil
}
