package scaleway

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/iot/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	defaultIoTRetryInterval = 5 * time.Second
	defaultIoTHubTimeout    = 5 * time.Minute
)

func iotAPIWithRegion(d *schema.ResourceData, m interface{}) (*iot.API, scw.Region, error) {
	meta := m.(*Meta)
	iotAPI := iot.NewAPI(meta.scwClient)

	region, err := extractRegion(d, meta)

	return iotAPI, region, err
}

func iotAPIWithRegionAndID(m interface{}, id string) (*iot.API, scw.Region, string, error) {
	meta := m.(*Meta)
	iotAPI := iot.NewAPI(meta.scwClient)

	region, ID, err := parseRegionalID(id)
	return iotAPI, region, ID, err
}

func waitIotHub(ctx context.Context, api *iot.API, region scw.Region, id string, timeout time.Duration) (*iot.Hub, error) {
	retryInterval := defaultIoTRetryInterval
	if DefaultWaitRetryInterval != nil {
		retryInterval = *DefaultWaitRetryInterval
	}

	hub, err := api.WaitForHub(&iot.WaitForHubRequest{
		HubID:         id,
		Region:        region,
		RetryInterval: &retryInterval,
		Timeout:       scw.TimeDurationPtr(timeout),
	}, scw.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("error waiting for IoT Hub (%s): %s", id, err)
	}

	return hub, nil
}

func extractRestHeaders(d *schema.ResourceData, key string) map[string]string {
	stringMap := map[string]string{}

	data := d.Get(key).(map[string]interface{})

	for k, v := range data {
		stringMap[k] = v.(string)
	}
	return stringMap
}
