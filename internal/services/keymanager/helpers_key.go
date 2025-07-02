package keymanager

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	key_manager "github.com/scaleway/scaleway-sdk-go/api/key_manager/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

const defaultKeyTimeout = 5 * time.Minute

func newKeyManagerAPI(d *schema.ResourceData, m any) (*key_manager.API, scw.Region, error) {
	api := key_manager.NewAPI(meta.ExtractScwClient(m))

	region, err := meta.ExtractRegion(d, m)
	if err != nil {
		return nil, "", err
	}

	return api, region, nil
}

func NewKeyManagerAPIWithRegionAndID(m any, regionalID string) (*key_manager.API, scw.Region, string, error) {
	api := key_manager.NewAPI(meta.ExtractScwClient(m))

	region, ID, err := regional.ParseID(regionalID)
	if err != nil {
		return nil, "", "", err
	}

	return api, region, ID, nil
}
