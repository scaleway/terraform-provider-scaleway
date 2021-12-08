package scaleway

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/vpc/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// vpcAPIWithZone returns a new VPC API and the zone for a Create request
func vpcAPIWithZone(d *schema.ResourceData, m interface{}) (*vpc.API, scw.Zone, error) {
	meta := m.(*Meta)
	vpcAPI := vpc.NewAPI(meta.scwClient)

	zone, err := extractZone(d, meta)
	if err != nil {
		return nil, "", err
	}
	return vpcAPI, zone, err
}

// vpcAPIWithZoneAndID
func vpcAPIWithZoneAndID(m interface{}, id string) (*vpc.API, scw.Zone, string, error) {
	meta := m.(*Meta)
	vpcAPI := vpc.NewAPI(meta.scwClient)

	zone, ID, err := parseZonedID(id)
	if err != nil {
		return nil, "", "", err
	}
	return vpcAPI, zone, ID, err
}

func vpcAPI(m interface{}) (*vpc.API, error) {
	meta, ok := m.(*Meta)
	if !ok {
		return nil, fmt.Errorf("wrong type: %T", m)
	}

	return vpc.NewAPI(meta.scwClient), nil
}

const minVPCTimeOut = 10 * time.Second
const PollVPCTimeOut = 30 * time.Second

// retryVPCContext is a basic wrapper around StateChangeConf that will just retry with a poll interval
// a function until it no longer returns an error.
//
// Cancellation from the passed in context will propagate through to the
// underlying StateChangeConf
func retryVPCContext(ctx context.Context, timeout time.Duration, f resource.RetryFunc) error {
	// These are used to pull the error out of the function; need a mutex to
	// avoid a data race.
	var resultErr error
	var resultErrMu sync.Mutex

	c := &resource.StateChangeConf{
		Pending:      []string{"retryableerror"},
		Target:       []string{"success"},
		Timeout:      timeout,
		MinTimeout:   minVPCTimeOut,
		PollInterval: PollVPCTimeOut,
		Refresh: func() (interface{}, string, error) {
			rErr := f()

			resultErrMu.Lock()
			defer resultErrMu.Unlock()

			if rErr == nil {
				resultErr = nil
				return 42, "success", nil
			}

			resultErr = rErr.Err

			if rErr.Retryable {
				return 42, "retryableerror", nil
			}
			return nil, "quit", rErr.Err
		},
	}

	_, waitErr := c.WaitForStateContext(ctx)

	// Need to acquire the lock here to be able to avoid race using resultErr as
	// the return value
	resultErrMu.Lock()
	defer resultErrMu.Unlock()

	// resultErr may be nil because the wait timed out and resultErr was never
	// set; this is still an error
	if resultErr == nil {
		return waitErr
	}
	// resultErr takes precedence over waitErr if both are set because it is
	// more likely to be useful
	return resultErr
}
