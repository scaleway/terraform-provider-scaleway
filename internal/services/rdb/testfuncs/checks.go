package rdbtestfuncs

import (
	"context"
	"fmt"
	"strings"
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

// testAccRDBListVCRProjectID is the default project ID in RDB list VCR cassettes.
const testAccRDBListVCRProjectID = "105bdce1-64c0-48ab-899d-868455867ecf"

// ListProjectID returns project_id for RDB list acceptance tests: SDK default when set,
// otherwise the VCR placeholder so replay matches committed cassettes.
func ListProjectID(tt *acctest.TestTools) string {
	pid, ok := tt.Meta.ScwClient().GetDefaultProjectID()
	if ok {
		if s := strings.TrimSpace(pid); s != "" {
			return s
		}
	}

	return testAccRDBListVCRProjectID
}

func HasNoPublicEndpoint(tt *acctest.TestTools, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}

		api, region, id, err := rdb.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		instance, err := api.GetInstance(&rdbSDK.GetInstanceRequest{
			Region:     region,
			InstanceID: id,
		})
		if err != nil {
			return fmt.Errorf("failed to get instance %s: %w", rs.Primary.ID, err)
		}

		for _, endpoint := range instance.Endpoints {
			if endpoint.LoadBalancer != nil {
				return fmt.Errorf(
					"instance %s has unexpected public endpoint %s (ip=%v, port=%d)",
					resourceName,
					endpoint.ID,
					endpoint.IP,
					endpoint.Port,
				)
			}
		}

		return nil
	}
}

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
