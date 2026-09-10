package messageq_test

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	messageqSDK "github.com/scaleway/scaleway-sdk-go/api/messageq/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/messageq"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

const deploymentTestUserName = "my_initial_user"

func fetchLatestVersion(tt *acctest.TestTools) string {
	tt.T.Helper()

	api := messageq.NewAPI(tt.Meta)

	versionsResp, err := api.ListVersions(&messageqSDK.ListVersionsRequest{
		Region: scw.RegionFrPar,
	}, scw.WithAllPages())
	if err != nil {
		tt.T.Fatalf("unable to fetch messageq versions: %s", err)
	}

	for _, version := range versionsResp.Versions {
		if !version.Disabled {
			return version.Version
		}
	}

	tt.T.Fatal("no messageq versions available")

	return ""
}

func fetchAvailableNodeType(tt *acctest.TestTools) string {
	tt.T.Helper()

	api := messageq.NewAPI(tt.Meta)

	nodeTypesResp, err := api.ListNodeTypes(&messageqSDK.ListNodeTypesRequest{
		Region: scw.RegionFrPar,
	}, scw.WithAllPages())
	if err != nil {
		tt.T.Fatalf("unable to fetch messageq node types: %s", err)
	}

	for _, nodeType := range nodeTypesResp.NodeTypes {
		if nodeType.StockStatus == messageqSDK.NodeTypeStockStatusAvailable && !nodeType.Disabled {
			return nodeType.Name
		}
	}

	tt.T.Fatal("no available messageq node types found")

	return ""
}

func isDeploymentPresent(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, id, err := messageq.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		err = transport.RetryOn403(tt.T.Context(), func() error {
			_, err = api.GetDeployment(&messageqSDK.GetDeploymentRequest{
				Region:       region,
				DeploymentID: id,
			}, scw.WithContext(tt.T.Context()))

			return err
		})

		return err
	}
}
