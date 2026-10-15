package kafka

import (
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/go-cty/cty"
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

// expandClusterAvailability maps multi_az / mono_az_zone to the API oneof.
// If neither is set, defaults to mono-AZ on the provider default zone.
func expandClusterAvailability(d *schema.ResourceData, m any) (*kafkaapi.ClusterMultiAZDetails, *kafkaapi.ClusterMonoAZDetails, error) {
	if raw, ok := meta.GetRawConfigForKey(d, "multi_az", cty.Bool); ok {
		if raw.(bool) {
			return &kafkaapi.ClusterMultiAZDetails{}, nil, nil
		}

		return nil, nil, errors.New("multi_az cannot be false; set mono_az_zone or omit multi_az to use the provider default zone")
	}

	if v, ok := d.GetOk("mono_az_zone"); ok {
		zone, err := scw.ParseZone(v.(string))
		if err != nil {
			return nil, nil, fmt.Errorf("invalid mono_az_zone: %w", err)
		}

		return nil, &kafkaapi.ClusterMonoAZDetails{Zone: &zone}, nil
	}

	zone, err := meta.ExtractZone(d, m)
	if err != nil {
		return nil, nil, fmt.Errorf("mono_az default zone: %w", err)
	}

	return nil, &kafkaapi.ClusterMonoAZDetails{Zone: &zone}, nil
}
