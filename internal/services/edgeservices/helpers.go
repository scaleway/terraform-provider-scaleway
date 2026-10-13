package edgeservices

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	edgeservices "github.com/scaleway/scaleway-sdk-go/api/edge_services/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

// NewEdgeServicesAPI returns a new edge_services API
func NewEdgeServicesAPI(m any) *edgeservices.API {
	return edgeservices.NewAPI(meta.ExtractScwClient(m))
}

// NewEdgeServicesAPIWithRegion returns a new edge_services API and the region
func NewEdgeServicesAPIWithRegion(d *schema.ResourceData, m any) (*edgeservices.API, scw.Region, error) {
	api := edgeservices.NewAPI(meta.ExtractScwClient(m))

	region, err := meta.ExtractRegion(d, m)
	if err != nil {
		return nil, "", err
	}

	return api, region, err
}

// edgeServicesAPIWithZoneAndRegion returns a new edge_services API, the zone and the region
func edgeServicesAPIWithZoneAndRegion(d *schema.ResourceData, m any) (*edgeservices.API, scw.Zone, scw.Region, error) {
	api := edgeservices.NewAPI(meta.ExtractScwClient(m))

	zone, err := meta.ExtractZone(d, m)
	if err != nil {
		return nil, "", "", err
	}

	region, err := meta.ExtractRegion(d, m)
	if err != nil {
		return nil, "", "", err
	}

	return api, zone, region, nil
}

var backendStageConfigKeys = []string{
	"s3_backend_config",
	"lb_backend_config",
	"container_backend_config",
	"function_backend_config",
}

func backendStageCustomizeDiff(_ context.Context, diff *schema.ResourceDiff, _ any) error {
	if diff.Id() == "" {
		return nil
	}

	var oldType, newType string

	for _, key := range backendStageConfigKeys {
		oldConfig, newConfig := diff.GetChange(key)
		if list, ok := oldConfig.([]any); ok && len(list) > 0 {
			oldType = key
		}

		if list, ok := newConfig.([]any); ok && len(list) > 0 {
			newType = key
		}
	}

	if oldType != "" && newType != "" && oldType != newType {
		return diff.ForceNew(newType)
	}

	return nil
}