package scaleway

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	flexibleip "github.com/scaleway/scaleway-sdk-go/api/flexibleip/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	defaultFlexibleIPTimeout = 1 * time.Minute
	retryFlexibleIPInterval  = 5 * time.Second
)

// fipAPIWithZone returns an lb API WITH zone for a Create request
func fipAPIWithZone(d *schema.ResourceData, m interface{}) (*flexibleip.API, scw.Zone, error) {
	meta := m.(*Meta)
	flexibleipAPI := flexibleip.NewAPI(meta.scwClient)

	zone, err := extractZone(d, meta)
	if err != nil {
		return nil, "", err
	}
	return flexibleipAPI, zone, nil
}

// fipAPIWithZoneAndID returns an flexibleip API with zone and ID extracted from the state
func fipAPIWithZoneAndID(m interface{}, id string) (*flexibleip.API, scw.Zone, string, error) {
	meta := m.(*Meta)
	fipAPI := flexibleip.NewAPI(meta.scwClient)

	zone, ID, err := parseZonedID(id)
	if err != nil {
		return nil, "", "", err
	}
	return fipAPI, zone, ID, nil
}

func waitFlexibleIP(ctx context.Context, api *flexibleip.API, zone scw.Zone, id string, timeout time.Duration) (*flexibleip.FlexibleIP, error) {
	retryInterval := retryFlexibleIPInterval
	if DefaultWaitRetryInterval != nil {
		retryInterval = *DefaultWaitRetryInterval
	}

	fip, err := api.WaitForFlexibleIP(&flexibleip.WaitForFlexibleIPRequest{
		FipID:         id,
		Zone:          zone,
		Timeout:       scw.TimeDurationPtr(timeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("error while waiting for flexible ip: %w", err)
	}

	return fip, err
}
