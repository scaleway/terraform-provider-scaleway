package autoscaling

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	autoscaling "github.com/scaleway/scaleway-sdk-go/api/autoscaling/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

// NewAPIWithZone returns a new autoscaling API and the zone for a Create request
func NewAPIWithZone(d *schema.ResourceData, m any) (*autoscaling.API, scw.Zone, error) {
	autoscalingAPI := autoscaling.NewAPI(meta.ExtractScwClient(m))

	zone, err := meta.ExtractZone(d, m)
	if err != nil {
		return nil, "", err
	}

	return autoscalingAPI, zone, nil
}

// NewAPIWithZoneAndID returns a new autoscaling API with zone and ID extracted from the state
func NewAPIWithZoneAndID(m any, zonalID string) (*autoscaling.API, scw.Zone, string, error) {
	autoscalingAPI := autoscaling.NewAPI(meta.ExtractScwClient(m))

	zone, ID, err := zonal.ParseID(zonalID)
	if err != nil {
		return nil, "", "", err
	}

	return autoscalingAPI, zone, ID, nil
}
