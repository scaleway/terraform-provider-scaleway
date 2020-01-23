package scaleway

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/registry/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func registryNamespaceWithRegion(d *schema.ResourceData, m interface{}) (*registry.API, scw.Region, error) {
	meta := m.(*Meta)
	api := registry.NewAPI(meta.scwClient)

	region, err := getRegion(d, meta)
	return api, region, err
}

func registryNamespaceWithRegionAndID(m interface{}, id string) (*registry.API, scw.Region, string, error) {
	meta := m.(*Meta)
	api := registry.NewAPI(meta.scwClient)

	region, id, err := parseRegionalID(id)
	return api, region, id, err
}
