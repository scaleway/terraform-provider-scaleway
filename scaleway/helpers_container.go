package scaleway

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	container "github.com/scaleway/scaleway-sdk-go/api/container/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	defaultContainerNamespaceTimeout = 20 * time.Second
)

// containerAPIWithRegion returns a new container API and the region.
func containerAPIWithRegion(d *schema.ResourceData, m interface{}) (*container.API, scw.Region, error) {
	meta := m.(*Meta)
	api := container.NewAPI(meta.scwClient)

	region, err := extractRegion(d, meta)
	if err != nil {
		return nil, "", err
	}
	return api, region, nil
}

// containerAPIWithRegionAndID returns a new container API, region and ID.
func containerAPIWithRegionAndID(m interface{}, id string) (*container.API, scw.Region, string, error) {
	meta := m.(*Meta)
	api := container.NewAPI(meta.scwClient)

	region, id, err := parseRegionalID(id)
	if err != nil {
		return nil, "", "", err
	}
	return api, region, id, nil
}
