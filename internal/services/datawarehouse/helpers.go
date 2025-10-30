package datawarehouse

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	datawarehouseapi "github.com/scaleway/scaleway-sdk-go/api/datawarehouse/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

const (
	defaultWaitRetryInterval = 30 * time.Second
)

func NewAPI(m any) *datawarehouseapi.API {
	return datawarehouseapi.NewAPI(meta.ExtractScwClient(m))
}

// datawarehouseAPIWithRegion returns a new Datawarehouse API and the region for a Create request
func datawarehouseAPIWithRegion(d *schema.ResourceData, m any) (*datawarehouseapi.API, scw.Region, error) {
	api := datawarehouseapi.NewAPI(meta.ExtractScwClient(m))

	region, err := meta.ExtractRegion(d, m)
	if err != nil {
		return nil, "", err
	}

	return api, region, nil
}

// NewAPIWithRegionAndID returns a Datawarehouse API with region and ID extracted from the state
func NewAPIWithRegionAndID(m any, id string) (*datawarehouseapi.API, scw.Region, string, error) {
	api := datawarehouseapi.NewAPI(meta.ExtractScwClient(m))

	region, id, err := regional.ParseID(id)
	if err != nil {
		return nil, "", "", err
	}

	return api, region, id, nil
}
