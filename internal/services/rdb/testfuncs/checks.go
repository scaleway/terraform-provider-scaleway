package rdbtestfuncs

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	rdbSDK "github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/rdb"
)

var DestroyWaitTimeout = 3 * time.Minute

func IsInstanceDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		ctx := context.Background()

		return retry.RetryContext(ctx, DestroyWaitTimeout, func() *retry.RetryError {
			for _, rs := range state.RootModule().Resources {
				if rs.Type != "scaleway_rdb_instance" {
					continue
				}

				api, region, id, err := rdb.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
				if err != nil {
					return retry.NonRetryableError(err)
				}

				_, err = api.GetInstance(&rdbSDK.GetInstanceRequest{
					InstanceID: id,
					Region:     region,
				})

				switch {
				case err == nil:
					return retry.RetryableError(fmt.Errorf("rdb instance (%s) still exists", rs.Primary.ID))
				case httperrors.Is404(err):
					continue
				default:
					return retry.NonRetryableError(err)
				}
			}

			return nil
		})
	}
}

func GetLatestEngineVersion(tt *acctest.TestTools, engineName string) string {
	api := rdbSDK.NewAPI(tt.Meta.ScwClient())

	engines, err := api.ListDatabaseEngines(&rdbSDK.ListDatabaseEnginesRequest{})
	if err != nil {
		tt.T.Fatalf("Could not get latest engine version: %s", err)
	}

	latestEngineVersion := ""

	for _, engine := range engines.Engines {
		if engine.Name == engineName {
			if len(engine.Versions) > 0 {
				latestEngineVersion = engine.Versions[0].Name

				break
			}
		}
	}

	return latestEngineVersion
}

func GetEngineVersionsForUpgrade(tt *acctest.TestTools, engineName string) (string, string) {
	api := rdbSDK.NewAPI(tt.Meta.ScwClient())

	engines, err := api.ListDatabaseEngines(&rdbSDK.ListDatabaseEnginesRequest{})
	if err != nil {
		tt.T.Fatalf("Could not get engine versions: %s", err)
	}

	for _, engine := range engines.Engines {
		if engine.Name == engineName {
			var availableVersions []string

			for _, version := range engine.Versions {
				if !version.Disabled {
					availableVersions = append(availableVersions, version.Name)
				}
			}

			if len(availableVersions) >= 2 {
				return availableVersions[1], availableVersions[0]
			}
		}
	}

	tt.T.Fatalf("Could not find two different versions for engine %s", engineName)

	return "", ""
}
