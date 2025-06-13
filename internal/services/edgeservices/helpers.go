package edgeservices

import (
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
func edgeServicesAPIWithZone(d *schema.ResourceData, m interface{}) (*edgeservices.API, scw.Zone, error) {
	api := edgeservices.NewAPI(meta.ExtractScwClient(m))

	zone, err := meta.ExtractZone(d, m)
	if err != nil {
		return nil, "", err
	}

	return api, zone, nil
}
