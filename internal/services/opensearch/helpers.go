package opensearch

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	searchdbapi "github.com/scaleway/scaleway-sdk-go/api/searchdb/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

const (
	defaultWaitRetryInterval = 30 * time.Second
)

func NewAPI(m any) *searchdbapi.API {
	return searchdbapi.NewAPI(meta.ExtractScwClient(m))
}

// opensearchAPIWithRegion returns a new OpenSearch API and the region for a Create request
func opensearchAPIWithRegion(d *schema.ResourceData, m any) (*searchdbapi.API, scw.Region, error) {
	api := searchdbapi.NewAPI(meta.ExtractScwClient(m))

	region, err := meta.ExtractRegion(d, m)
	if err != nil {
		return nil, "", err
	}

	return api, region, nil
}

// NewAPIWithRegionAndID returns an OpenSearch API with region and ID extracted from the state
func NewAPIWithRegionAndID(m any, id string) (*searchdbapi.API, scw.Region, string, error) {
	api := searchdbapi.NewAPI(meta.ExtractScwClient(m))

	region, id, err := regional.ParseID(id)
	if err != nil {
		return nil, "", "", err
	}

	return api, region, id, nil
}
