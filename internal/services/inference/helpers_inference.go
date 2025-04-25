package inference

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/inference/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

const (
	defaultInferenceDeploymentTimeout = 80 * time.Minute
	defaultDeploymentRetryInterval    = 1 * time.Minute
	defaultCustomModelTimeout         = 40 * time.Minute
	defaultCustomModelRetryInterval   = 1 * time.Minute
)

// NewAPIWithRegion returns a new inference API and the region for a Create request
func NewAPIWithRegion(d *schema.ResourceData, m interface{}) (*inference.API, scw.Region, error) {
	inferenceAPI := inference.NewAPI(meta.ExtractScwClient(m))

	region, err := meta.ExtractRegion(d, m)
	if err != nil {
		return nil, "", err
	}

	return inferenceAPI, region, nil
}

// NewAPIWithRegionAndID returns a new inference API with region and ID extracted from the state
func NewAPIWithRegionAndID(m interface{}, regionalID string) (*inference.API, scw.Region, string, error) {
	inferenceAPI := inference.NewAPI(meta.ExtractScwClient(m))

	region, ID, err := regional.ParseID(regionalID)
	if err != nil {
		return nil, "", "", err
	}

	return inferenceAPI, region, ID, nil
}
