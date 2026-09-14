package messageq

import (
	"time"

	messageqapi "github.com/scaleway/scaleway-sdk-go/api/messageq/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

const (
	defaultWaitRetryInterval     = 30 * time.Second
	defaultDeploymentTimeout     = 30 * time.Minute
	defaultDeploymentReadTimeout = 5 * time.Minute
)

func NewAPI(m any) *messageqapi.API {
	return messageqapi.NewAPI(meta.ExtractScwClient(m))
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

// RegionAndIDFromAttr resolves the region and the bare UUID from an attribute that
// accepts either a plain UUID or a localized "{region}/{uuid}" value. The fallback
// region is kept when the attribute carries no locality.
func RegionAndIDFromAttr(value string, fallback scw.Region) (scw.Region, string) {
	region, id, err := regional.ParseID(value)
	if err != nil {
		return fallback, locality.ExpandID(value)
	}

	return region, id
}

func ExpandVolumeSizeBytes(sizeInGB int) scw.Size {
	return scw.Size(sizeInGB) * scw.GB
}

// BytesToGB converts an API size to decimal GB. The MessageQ API expresses both
// volume and memory sizes in decimal bytes, so the same conversion applies to both.
func BytesToGB(sizeBytes scw.Size) int {
	return int(sizeBytes / scw.GB)
}
