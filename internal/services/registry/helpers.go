package registry

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/registry/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

const (
	defaultNamespaceTimeout       = 5 * time.Minute
	defaultNamespaceRetryInterval = 5 * time.Second
)

type ErrorRegistryMessage struct {
	Error string `json:"error"`
}

// NewAPIWithRegion returns a new container registry API and the region.
func NewAPIWithRegion(d *schema.ResourceData, m any) (*registry.API, scw.Region, error) {
	api := registry.NewAPI(meta.ExtractScwClient(m))

	region, err := meta.ExtractRegion(d, m)
	if err != nil {
		return nil, "", err
	}

	return api, region, nil
}

// NewAPIWithRegionAndID returns a new container registry API, region and ID.
func NewAPIWithRegionAndID(m any, id string) (*registry.API, scw.Region, string, error) {
	api := registry.NewAPI(meta.ExtractScwClient(m))

	region, id, err := regional.ParseID(id)
	if err != nil {
		return nil, "", "", err
	}

	return api, region, id, nil
}
