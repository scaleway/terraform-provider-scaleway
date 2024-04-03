package block

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	block "github.com/scaleway/scaleway-sdk-go/api/block/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

const (
	defaultBlockTimeout       = 5 * time.Minute
	defaultBlockRetryInterval = 5 * time.Second
	BlockVolumeType           = instance.VolumeServerVolumeType("sbs_volume")
)

// blockAPIWithZone returns a new block API and the zone for a Create request
func blockAPIWithZone(d *schema.ResourceData, m interface{}) (*block.API, scw.Zone, error) {
	blockAPI := block.NewAPI(meta.ExtractScwClient(m))

	zone, err := meta.ExtractZone(d, m)
	if err != nil {
		return nil, "", err
	}

	return blockAPI, zone, nil
}

// NewAPIWithZoneAndID returns a new block API with zone and ID extracted from the state
func NewAPIWithZoneAndID(m interface{}, zonedID string) (*block.API, scw.Zone, string, error) {
	blockAPI := block.NewAPI(meta.ExtractScwClient(m))

	zone, ID, err := zonal.ParseID(zonedID)
	if err != nil {
		return nil, "", "", err
	}

	return blockAPI, zone, ID, nil
}

// customDiffCannotShrink set key to forceNew if value shrink
func customDiffCannotShrink(key string) schema.CustomizeDiffFunc {
	return customdiff.ForceNewIf(key, func(_ context.Context, d *schema.ResourceDiff, _ interface{}) bool {
		oldValueI, newValueI := d.GetChange(key)
		oldValue := oldValueI.(int)
		newValue := newValueI.(int)

		return oldValue < newValue
	})
}
