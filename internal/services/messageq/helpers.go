package messageq

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	messageqapi "github.com/scaleway/scaleway-sdk-go/api/messageq/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

const (
	defaultWaitRetryInterval = 30 * time.Second
	defaultDeploymentTimeout = 30 * time.Minute
)

func NewAPI(m any) *messageqapi.API {
	return messageqapi.NewAPI(meta.ExtractScwClient(m))
}

// newAPIWithRegion returns a new MessageQ API and the region for a Create request.
func newAPIWithRegion(d *schema.ResourceData, m any) (*messageqapi.API, scw.Region, error) {
	api := messageqapi.NewAPI(meta.ExtractScwClient(m))

	region, err := meta.ExtractRegion(d, m)
	if err != nil {
		return nil, "", err
	}

	return api, region, nil
}

// NewAPIWithRegionAndID returns a MessageQ API with region and ID extracted from the state.
func NewAPIWithRegionAndID(m any, id string) (*messageqapi.API, scw.Region, string, error) {
	api := messageqapi.NewAPI(meta.ExtractScwClient(m))

	region, id, err := regional.ParseID(id)
	if err != nil {
		return nil, "", "", err
	}

	return api, region, id, nil
}

func expandVolumeSizeBytes(sizeInGB int) scw.Size {
	return scw.Size(uint64(sizeInGB) * 1000 * 1000 * 1000)
}

func flattenVolumeSizeGB(sizeBytes scw.Size) int {
	return int(uint64(sizeBytes) / (1000 * 1000 * 1000))
}
