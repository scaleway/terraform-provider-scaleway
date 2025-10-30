package datawarehouse_test

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	datawarehouseSDK "github.com/scaleway/scaleway-sdk-go/api/datawarehouse/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/datawarehouse"
)

// Common helper functions shared across all datawarehouse tests

// fetchLatestClickHouseVersion returns the latest available ClickHouse version for testing purposes
func fetchLatestClickHouseVersion(tt *acctest.TestTools) string {
	tt.T.Helper()

	api := datawarehouse.NewAPI(tt.Meta)

	versionsResp, err := api.ListVersions(&datawarehouseSDK.ListVersionsRequest{}, scw.WithAllPages())
	if err != nil {
		tt.T.Fatalf("unable to fetch datawarehouse versions: %s", err)
	}

	if len(versionsResp.Versions) == 0 {
		tt.T.Fatal("no datawarehouse versions available")
	}

	return versionsResp.Versions[0].Version
}

func isDeploymentPresent(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, id, err := datawarehouse.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetDeployment(&datawarehouseSDK.GetDeploymentRequest{
			Region:       region,
			DeploymentID: id,
		}, scw.WithContext(context.Background()))

		return err
	}
}
