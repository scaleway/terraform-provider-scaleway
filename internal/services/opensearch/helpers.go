package opensearch

import (
	"context"
	"fmt"
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

func waitForDeploymentEndpointState(
	ctx context.Context,
	api *searchdbapi.API,
	region scw.Region,
	deploymentID string,
	timeout time.Duration,
	desiredPrivate bool,
	privateNetworkID string,
) error {
	deadline := time.Now().Add(timeout)

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for deployment %s endpoints to match desired connectivity", deploymentID)
		}

		deployment, err := api.GetDeployment(&searchdbapi.GetDeploymentRequest{
			Region:       region,
			DeploymentID: deploymentID,
		}, scw.WithContext(ctx))
		if err != nil {
			return err
		}

		if deploymentEndpointStateMatches(deployment, desiredPrivate, privateNetworkID) {
			return nil
		}

		time.Sleep(defaultWaitRetryInterval)
	}
}

func deploymentEndpointStateMatches(
	deployment *searchdbapi.Deployment,
	desiredPrivate bool,
	privateNetworkID string,
) bool {
	var hasPrivate bool

	privateNetworkIDMatches := false

	for _, endpoint := range deployment.Endpoints {
		if endpoint == nil {
			continue
		}

		if endpoint.PrivateNetwork != nil {
			hasPrivate = true

			if endpoint.PrivateNetwork.PrivateNetworkID == privateNetworkID {
				privateNetworkIDMatches = true
			}
		}
	}

	if desiredPrivate {
		// A public dashboard endpoint may remain while the private API endpoint is provisioned.
		return hasPrivate && privateNetworkIDMatches
	}

	return !hasPrivate
}
