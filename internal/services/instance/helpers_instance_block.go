package instance

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	block "github.com/scaleway/scaleway-sdk-go/api/block/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

type BlockAndInstanceAPI struct {
	*instance.API
	blockAPI *block.API
}

// newAPIWithZone returns a new instance API and the zone for a Create request
func instanceAndBlockAPIWithZone(d *schema.ResourceData, m interface{}) (*BlockAndInstanceAPI, scw.Zone, error) {
	instanceAPI := instance.NewAPI(meta.ExtractScwClient(m))
	blockAPI := block.NewAPI(meta.ExtractScwClient(m))

	zone, err := meta.ExtractZone(d, m)
	if err != nil {
		return nil, "", err
	}

	return &BlockAndInstanceAPI{
		API:      instanceAPI,
		blockAPI: blockAPI,
	}, zone, nil
}

// NewAPIWithZoneAndID returns an instance API with zone and ID extracted from the state
func instanceAndBlockAPIWithZoneAndID(m interface{}, zonedID string) (*BlockAndInstanceAPI, scw.Zone, string, error) {
	instanceAPI := instance.NewAPI(meta.ExtractScwClient(m))
	blockAPI := block.NewAPI(meta.ExtractScwClient(m))

	zone, ID, err := zonal.ParseID(zonedID)
	if err != nil {
		return nil, "", "", err
	}

	return &BlockAndInstanceAPI{
		API:      instanceAPI,
		blockAPI: blockAPI,
	}, zone, ID, nil
}
