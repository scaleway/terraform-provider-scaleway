package edgeservices

import (
	"context"
	"fmt"

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

// edgeServicesAPIWithZone returns a new edge_services API and the zone
func edgeServicesAPIWithZone(d *schema.ResourceData, m any) (*edgeservices.API, scw.Zone, error) {
	api := edgeservices.NewAPI(meta.ExtractScwClient(m))

	zone, err := meta.ExtractZone(d, m)
	if err != nil {
		return nil, "", err
	}

	return api, zone, nil
}

func validateWAFBackendStageIsNotS3(ctx context.Context, api *edgeservices.API, backendStageID string) error {
	if backendStageID == "" {
		return nil
	}

	backendStage, err := api.GetBackendStage(&edgeservices.GetBackendStageRequest{
		BackendStageID: backendStageID,
	}, scw.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to get backend stage: %w", err)
	}

	if backendStage.ScalewayS3 != nil {
		return fmt.Errorf("WAF stage is only supported with Load Balancer backends")
	}

	return nil
}
