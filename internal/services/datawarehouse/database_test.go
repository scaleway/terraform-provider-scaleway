package datawarehouse_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	datawarehouseSDK "github.com/scaleway/scaleway-sdk-go/api/datawarehouse/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/datawarehouse"
)

func TestAccDatabase_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	latestVersion := fetchLatestClickHouseVersion(tt)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isDatabaseDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "scaleway_datawarehouse_deployment" "main" {
  name           = "tf-test-deploy-db"
  version        = "%s"
  replica_count  = 1
  cpu_min        = 2
  cpu_max        = 4
  ram_per_cpu    = 4
  password       = "password@1234567"
}

resource "scaleway_datawarehouse_database" "mydb" {
  deployment_id = scaleway_datawarehouse_deployment.main.id
  name          = "testdb"
}
`, latestVersion),
				Check: resource.ComposeTestCheckFunc(
					isDeploymentPresent(tt, "scaleway_datawarehouse_deployment.main"),
					isDatabasePresent(tt, "scaleway_datawarehouse_database.mydb"),
					resource.TestCheckResourceAttr("scaleway_datawarehouse_database.mydb", "name", "testdb"),
				),
			},
		},
	})
}

func isDatabasePresent(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		id := rs.Primary.ID // format: region/deployment_id/name

		region, deploymentID, dbName, err := datawarehouse.ResourceDatabaseParseID(id)
		if err != nil {
			return fmt.Errorf("unexpected ID format (%s), expected region/deployment_id/name", id)
		}

		api := datawarehouse.NewAPI(tt.Meta)

		resp, err := api.ListDatabases(&datawarehouseSDK.ListDatabasesRequest{
			Region:       region,
			DeploymentID: deploymentID,
			Name:         scw.StringPtr(dbName),
		}, scw.WithContext(context.Background()))
		if err != nil {
			return err
		}

		for _, db := range resp.Databases {
			if db.Name == dbName {
				return nil
			}
		}

		return fmt.Errorf("database %s not found", dbName)
	}
}

func isDatabaseDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "scaleway_datawarehouse_database" {
				continue
			}

			id := rs.Primary.ID // format: region/deployment_id/name

			region, deploymentID, dbName, err := datawarehouse.ResourceDatabaseParseID(id)
			if err != nil {
				return fmt.Errorf("unexpected ID format (%s), expected region/deployment_id/name", id)
			}

			api := datawarehouse.NewAPI(tt.Meta)

			resp, err := api.ListDatabases(&datawarehouseSDK.ListDatabasesRequest{
				Region:       region,
				DeploymentID: deploymentID,
				Name:         scw.StringPtr(dbName),
			}, scw.WithContext(context.Background()))
			if err != nil {
				if httperrors.Is404(err) {
					continue
				}

				return err
			}

			for _, db := range resp.Databases {
				if db.Name == dbName {
					return fmt.Errorf("database %s still exists", dbName)
				}
			}
		}

		return nil
	}
}
