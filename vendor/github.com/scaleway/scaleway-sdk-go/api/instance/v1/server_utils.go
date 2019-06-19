package instance

import (
	"time"

	"github.com/scaleway/scaleway-sdk-go/internal/async"
	"github.com/scaleway/scaleway-sdk-go/internal/errors"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/scaleway-sdk-go/utils"
)

// waitForServerRequest is used by waitForServer method
type waitForServerRequest struct {
	ServerID string
	Zone     utils.Zone
	Timeout  time.Duration
}

// waitForServer wait for the server to be in a "terminal state" before returning.
// This function can be used to wait for a server to be started for example.
func (s *API) waitForServer(req *waitForServerRequest) (*Server, scw.SdkError) {

	terminalStatus := map[ServerState]struct{}{
		ServerStateStopped:        {},
		ServerStateStoppedInPlace: {},
		ServerStateLocked:         {},
		ServerStateRunning:        {},
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
			_, isTerminal := terminalStatus[res.Server.State]

			return res.Server, err, isTerminal
		},
		Timeout:          req.Timeout,
		IntervalStrategy: async.LinearIntervalStrategy(5 * time.Second),
	})
	if err != nil {
		return nil, errors.Wrap(err, "waiting for server failed")
	}
	return server.(*Server), nil
}

// ServerActionAndWaitRequest is used by ServerActionAndWait method
type ServerActionAndWaitRequest struct {
	ServerID string
	Zone     utils.Zone
	Action   ServerAction

	// Timeout: maximum time to wait before (default: 5 minutes)
	Timeout time.Duration
}

// ServerActionAndWait start an action and wait for the server to be in the correct "terminal state"
// expected by this action.
func (s *API) ServerActionAndWait(req *ServerActionAndWaitRequest) error {
	if req.Timeout == 0 {
		req.Timeout = 5 * time.Minute
	}

	_, err := s.ServerAction(&ServerActionRequest{
		Zone:     req.Zone,
		ServerID: req.ServerID,
		Action:   req.Action,
	})
	if err != nil {
		return err
	}

	finalServer, err := s.waitForServer(&waitForServerRequest{
		Zone:     req.Zone,
		ServerID: req.ServerID,
		Timeout:  req.Timeout,
	})
	if err != nil {
		return err
	}

	// check the action was properly executed
	expectedState := ServerState("unknown")
	switch req.Action {
	case ServerActionPoweron, ServerActionReboot:
		expectedState = ServerStateRunning
	case ServerActionPoweroff:
		expectedState = ServerStateStopped
	case ServerActionStopInPlace:
		expectedState = ServerStateStoppedInPlace
	}

	// backup can be performed from any state
	if expectedState != ServerState("unknown") && finalServer.State != expectedState {
		return errors.New("expected state %s but found %s: %s", expectedState, finalServer.State, finalServer.StateDetail)
	}

	return nil
}

// GetServerTypeRequest is used by GetServerType
type GetServerTypeRequest struct {
	Zone utils.Zone
	Name string
}

// GetServerType get server type info by it's name.
func (s *API) GetServerType(req *GetServerTypeRequest) (*ServerType, error) {

	res, err := s.ListServersTypes(&ListServersTypesRequest{
		Zone: req.Zone,
	}, scw.WithAllPages())

	if err != nil {
		return nil, err
	}

	if serverType, exist := res.Servers[req.Name]; exist {
		return serverType, nil
	}

	return nil, errors.New("could not find server type %q", req.Name)
}
