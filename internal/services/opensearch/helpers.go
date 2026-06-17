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

// newAPIWithRegion returns a new OpenSearch API and the region for a Create request
func newAPIWithRegion(d *schema.ResourceData, m any) (*searchdbapi.API, scw.Region, error) {
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

func deploymentNodeCountFromConfig(d *schema.ResourceData) int {
	if value, ok := d.GetOk("node_count"); ok {
		return value.(int)
	}

	if value, ok := d.GetOk("node_amount"); ok {
		return value.(int)
	}

	return 0
}

func deploymentNodeCountForState(d *schema.ResourceData, deployment *searchdbapi.Deployment) int {
	if deployment.NodeCount != 0 {
		return int(deployment.NodeCount)
	}

	// The API may return 0 for node_count while the deployment is ready (notably on shared tiers).
	// Preserve the configured value to avoid spurious ForceNew drift.
	for _, key := range []string{"node_count", "node_amount"} {
		if value, ok := d.GetOk(key); ok {
			if configured := value.(int); configured != 0 {
				return configured
			}
		}
	}

	return 0
}

func setDeploymentNodeCountState(d *schema.ResourceData, deployment *searchdbapi.Deployment) {
	nodeCount := deploymentNodeCountForState(d, deployment)

	if _, ok := d.GetOk("node_amount"); ok {
		_ = d.Set("node_amount", nodeCount)

		return
	}

	_ = d.Set("node_count", nodeCount)
}
