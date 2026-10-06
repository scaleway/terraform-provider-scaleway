package opensearch_test

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	searchdbSDK "github.com/scaleway/scaleway-sdk-go/api/searchdb/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/opensearch"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

const deploymentTestUserName = "my_initial_user"

func fetchLatestVersion(tt *acctest.TestTools) string {
	tt.T.Helper()

	api := opensearch.NewAPI(tt.Meta)

	versionsResp, err := api.ListVersions(&searchdbSDK.ListVersionsRequest{
		Region: scw.RegionFrPar,
	}, scw.WithAllPages())
	if err != nil {
		tt.T.Fatalf("unable to fetch opensearch versions: %s", err)
	}

	if len(versionsResp.Versions) == 0 {
		tt.T.Fatal("no opensearch versions available")
	}

	return versionsResp.Versions[0].Version
}

// nodeTypeUsable reports whether a node type can be provisioned right now.
// SearchDB stocks node types as available or low_stock; out_of_stock types
// cannot be ordered.
func nodeTypeUsable(nodeType *searchdbSDK.NodeType) bool {
	return !nodeType.Disabled && nodeType.StockStatus != searchdbSDK.NodeTypeStockStatusOutOfStock
}

func fetchAvailableNodeType(tt *acctest.TestTools) string {
	tt.T.Helper()

	api := opensearch.NewAPI(tt.Meta)

	nodeTypesResp, err := api.ListNodeTypes(&searchdbSDK.ListNodeTypesRequest{
		Region: scw.RegionFrPar,
	}, scw.WithAllPages())
	if err != nil {
		tt.T.Fatalf("unable to fetch opensearch node types: %s", err)
	}

	for _, nodeType := range nodeTypesResp.NodeTypes {
		if nodeTypeUsable(nodeType) {
			return nodeType.Name
		}
	}

	tt.T.Fatal("no available opensearch node types found")

	return ""
}

// fetchAvailableDedicatedNodeType returns the first usable dedicated node type.
// Scaling node_count in place (UpgradeDeployment) only applies to dedicated
// node types: shared tiers are limited to a single node and return node_count
// = 0 from the API.
func fetchAvailableDedicatedNodeType(tt *acctest.TestTools) string {
	tt.T.Helper()

	api := opensearch.NewAPI(tt.Meta)

	nodeTypesResp, err := api.ListNodeTypes(&searchdbSDK.ListNodeTypesRequest{
		Region: scw.RegionFrPar,
	}, scw.WithAllPages())
	if err != nil {
		tt.T.Fatalf("unable to fetch opensearch node types: %s", err)
	}

	for _, nodeType := range nodeTypesResp.NodeTypes {
		if nodeTypeUsable(nodeType) && nodeType.InstanceRange == "DEDICATED" {
			return nodeType.Name
		}
	}

	tt.T.Fatal("no available dedicated opensearch node types found")

	return ""
}

func isDeploymentPresent(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, id, err := opensearch.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		err = transport.RetryOn403(tt.T.Context(), func() error {
			_, err = api.GetDeployment(&searchdbSDK.GetDeploymentRequest{
				Region:       region,
				DeploymentID: id,
			}, scw.WithContext(tt.T.Context()))

			return err
		})

		return err
	}
}
