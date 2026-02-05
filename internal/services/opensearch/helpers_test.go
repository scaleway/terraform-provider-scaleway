package opensearch_test

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	searchdbSDK "github.com/scaleway/scaleway-sdk-go/api/searchdb/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/opensearch"
)

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
		if nodeType.StockStatus == searchdbSDK.NodeTypeStockStatusAvailable && !nodeType.Disabled {
			return nodeType.Name
		}
	}

	tt.T.Fatal("no available opensearch node types found")

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

		_, err = api.GetDeployment(&searchdbSDK.GetDeploymentRequest{
			Region:       region,
			DeploymentID: id,
		}, scw.WithContext(context.Background()))

		return err
	}
}
