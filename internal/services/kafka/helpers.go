package kafka

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	kafkaapi "github.com/scaleway/scaleway-sdk-go/api/kafka/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

const (
	defaultWaitRetryInterval = 30 * time.Second
)

func NewAPI(m any) *kafkaapi.API {
	return kafkaapi.NewAPI(meta.ExtractScwClient(m))
}

// newAPIWithRegion returns a new Kafka API and the region for a Create request
func newAPIWithRegion(d *schema.ResourceData, m any) (*kafkaapi.API, scw.Region, error) {
	api := kafkaapi.NewAPI(meta.ExtractScwClient(m))

	region, err := meta.ExtractRegion(d, m)
	if err != nil {
		return nil, "", err
	}

	return api, region, nil
}

// NewAPIWithRegionAndID returns a Kafka API with region and ID extracted from the state
func NewAPIWithRegionAndID(m any, id string) (*kafkaapi.API, scw.Region, string, error) {
	api := kafkaapi.NewAPI(meta.ExtractScwClient(m))

	region, id, err := regional.ParseID(id)
	if err != nil {
		return nil, "", "", err
	}

	return api, region, id, nil
}
