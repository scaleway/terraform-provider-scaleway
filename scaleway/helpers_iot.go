package scaleway

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/iot/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

const (
	defaultIoTRetryInterval = 5 * time.Second
	defaultIoTHubTimeout    = 5 * time.Minute
	mqttCaURLDownload       = "https://iot.s3.nl-ams.scw.cloud/certificates/"
	mqttCaFileName          = "iot-hub-ca.pem"
)

func iotAPIWithRegion(d *schema.ResourceData, m interface{}) (*iot.API, scw.Region, error) {
	iotAPI := iot.NewAPI(meta.ExtractScwClient(m))

	region, err := meta.ExtractRegion(d, m)

	return iotAPI, region, err
}

func IotAPIWithRegionAndID(m interface{}, id string) (*iot.API, scw.Region, string, error) {
	iotAPI := iot.NewAPI(meta.ExtractScwClient(m))

	region, ID, err := regional.ParseID(id)
	return iotAPI, region, ID, err
}

func waitIotHub(ctx context.Context, api *iot.API, region scw.Region, id string, timeout time.Duration) (*iot.Hub, error) {
	retryInterval := defaultIoTRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	hub, err := api.WaitForHub(&iot.WaitForHubRequest{
		HubID:         id,
		Region:        region,
		RetryInterval: &retryInterval,
		Timeout:       scw.TimeDurationPtr(timeout),
	}, scw.WithContext(ctx))

	return hub, err
}

func extractRestHeaders(d *schema.ResourceData, key string) map[string]string {
	stringMap := map[string]string{}

	data := d.Get(key).(map[string]interface{})

	for k, v := range data {
		stringMap[k] = v.(string)
	}
	return stringMap
}

func computeIotHubCaURL(productPlan iot.HubProductPlan, region scw.Region) string {
	if productPlan == "plan_shared" || productPlan == "plan_unknown" {
		return ""
	}
	return mqttCaURLDownload + string(region) + "/" + mqttCaFileName
}

func computeIotHubMQTTCa(ctx context.Context, mqttCaURL string, m interface{}) (string, error) {
	if mqttCaURL == "" {
		return "", nil
	}
	var mqttCa *http.Response
	req, _ := http.NewRequestWithContext(ctx, "GET", mqttCaURL, nil)
	mqttCa, err := meta.ExtractHTTPClient(m).Do(req)
	if err != nil {
		return "", err
	}
	defer mqttCa.Body.Close()
	resp, _ := io.ReadAll(mqttCa.Body)
	return string(resp), nil
}
