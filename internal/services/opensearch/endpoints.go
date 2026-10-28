package opensearch

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	searchdbapi "github.com/scaleway/scaleway-sdk-go/api/searchdb/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
)

// privateNetworkDesireFromConfig returns whether a private network is configured
// and its expanded UUID.
func privateNetworkDesireFromConfig(d *schema.ResourceData) (desiredPrivate bool, pnID string) {
	v, ok := d.GetOk("private_network")
	if !ok {
		return false, ""
	}

	pnList := v.([]any)
	if len(pnList) == 0 {
		return false, ""
	}

	pnMap := pnList[0].(map[string]any)

	return true, locality.ExpandID(pnMap["private_network_id"].(string))
}

func endpointSpec(desiredPrivate bool, pnID string) *searchdbapi.EndpointSpec {
	if desiredPrivate {
		return &searchdbapi.EndpointSpec{
			PrivateNetwork: &searchdbapi.EndpointSpecPrivateNetworkDetails{
				PrivateNetworkID: pnID,
			},
		}
	}

	return &searchdbapi.EndpointSpec{
		Public: &searchdbapi.EndpointSpecPublicDetails{},
	}
}

// reconcileDeploymentEndpoints aligns deployment endpoints with the Terraform desire:
// either one private API endpoint on pnID, or a public endpoint that exposes
// api-opensearch. Public dashboard-only endpoints are never deleted (SearchDB
// ignores or never completes those deletes).
func reconcileDeploymentEndpoints(
	ctx context.Context,
	api *searchdbapi.API,
	region scw.Region,
	deploymentID string,
	endpoints []*searchdbapi.Endpoint,
	desiredPrivate bool,
	pnID string,
	timeout time.Duration,
) error {
	var deletedEndpointIDs []string

	hasDesiredEndpoint := false

	for _, endpoint := range endpoints {
		if endpoint == nil {
			continue
		}

		if endpoint.Public != nil && endpoint.PrivateNetwork == nil {
			if !desiredPrivate && EndpointHasOpenSearchAPI(endpoint) {
				hasDesiredEndpoint = true
			}

			continue
		}

		if endpoint.PrivateNetwork == nil {
			continue
		}

		if desiredPrivate && endpoint.PrivateNetwork.PrivateNetworkID == pnID {
			hasDesiredEndpoint = true

			continue
		}

		err := api.DeleteEndpoint(&searchdbapi.DeleteEndpointRequest{
			Region:     region,
			EndpointID: endpoint.ID,
		}, scw.WithContext(ctx))
		if err != nil {
			return err
		}

		deletedEndpointIDs = append(deletedEndpointIDs, endpoint.ID)
	}

	// DeleteEndpoint returns 204 immediately but removal is asynchronous.
	// waitForDeployment only checks top-level status (stays "ready"), so poll
	// until deleted endpoints are gone before creating new ones.
	if len(deletedEndpointIDs) > 0 {
		err := waitForEndpointsDeleted(ctx, api, region, deploymentID, deletedEndpointIDs, timeout)
		if err != nil {
			return err
		}
	}

	if hasDesiredEndpoint {
		return nil
	}

	_, err := api.CreateEndpoint(&searchdbapi.CreateEndpointRequest{
		Region:       region,
		DeploymentID: deploymentID,
		EndpointSpec: endpointSpec(desiredPrivate, pnID),
	}, scw.WithContext(ctx))
	if err != nil {
		return err
	}

	_, err = waitForDeployment(ctx, api, region, deploymentID, timeout)

	return err
}
