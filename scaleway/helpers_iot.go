package scaleway

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	iot "github.com/scaleway/scaleway-sdk-go/api/iot/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
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

func waitIotHub(iotAPI *iot.API, region scw.Region, hubID string, timeout time.Duration, desiredStates ...iot.HubStatus) error {
	hub, err := iotAPI.WaitForHub(&iot.WaitForHubRequest{
		HubID:         hubID,
		Region:        region,
		RetryInterval: DefaultWaitRetryInterval,
		Timeout:       scw.TimeDurationPtr(timeout),
	})
	if err != nil {
		return err
	}

	for _, desiredState := range desiredStates {
		if hub.Status == desiredState {
			return nil
		}
	}

	return fmt.Errorf("hub %s has state %s, wants one of %+q", hubID, hub.Status, desiredStates)
}

func extractRestHeaders(d *schema.ResourceData, key string) map[string]string {
	stringMap := map[string]string{}

	data := d.Get(key).(map[string]interface{})

	for k, v := range data {
		stringMap[k] = v.(string)
	}
	return stringMap
}
