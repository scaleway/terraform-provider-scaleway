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

func isDeploymentPresent(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		id := rs.Primary.ID
		region := rs.Primary.Attributes["region"]

		api := datawarehouse.NewAPI(tt.Meta)
		_, err := api.GetDeployment(&datawarehouseSDK.GetDeploymentRequest{
			Region:       scw.Region(region),
			DeploymentID: id,
		}, scw.WithContext(context.Background()))

		return err
	}
}
