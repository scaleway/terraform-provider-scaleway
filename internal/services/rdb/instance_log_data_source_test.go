package rdb_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	rdbSDK "github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	rdbchecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/rdb/testfuncs"
)

func TestAccDataSourceInstanceLog_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	latestEngineVersion := rdbchecks.GetLatestEngineVersion(tt, postgreSQLEngineName)

	instanceConfig := fmt.Sprintf(`
		resource "scaleway_rdb_instance" "main" {
			name = "tf-tests-rdb-instance-log"
			engine = %q
			node_type = "db-dev-s"
			is_ha_cluster = false
			disable_backup = true
			user_name = "my_initial_user"
			password = "thiZ_is_v&ry_s3cret"
		}
	`, latestEngineVersion)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             rdbchecks.IsInstanceDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: instanceConfig,
				Check: resource.ComposeTestCheckFunc(
					isInstancePresent(tt, "scaleway_rdb_instance.main"),
					prepareInstanceLogs(tt, "scaleway_rdb_instance.main"),
				),
			},
			{
				Config: instanceConfig + `
					data "scaleway_rdb_instance_logs" "main" {
						instance_id = scaleway_rdb_instance.main.id
						order_by = "created_at_desc"
					}

					data "scaleway_rdb_instance_log" "main" {
						instance_log_id = data.scaleway_rdb_instance_logs.main.instance_logs[0].id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scaleway_rdb_instance_logs.main", "instance_logs.0.id"),
					resource.TestCheckResourceAttrSet("data.scaleway_rdb_instance_log.main", "id"),
					resource.TestCheckResourceAttrSet("data.scaleway_rdb_instance_log.main", "status"),
					resource.TestCheckResourceAttrSet("data.scaleway_rdb_instance_log.main", "node_name"),
					resource.TestCheckResourceAttrSet("data.scaleway_rdb_instance_log.main", "region"),
				),
			},
		},
	})
}

func prepareInstanceLogs(tt *acctest.TestTools, instanceResourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[instanceResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", instanceResourceName)
		}

		region, instanceID, err := regional.ParseID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("failed to parse instance ID: %w", err)
		}

		api := rdbSDK.NewAPI(tt.Meta.ScwClient())
		ctx := tt.T.Context()

		_, err = api.PrepareInstanceLogs(&rdbSDK.PrepareInstanceLogsRequest{
			Region:     region,
			InstanceID: instanceID,
		}, scw.WithContext(ctx))
		if err != nil {
			return fmt.Errorf("failed to prepare instance logs: %w", err)
		}

		return retry.RetryContext(ctx, 5*time.Minute, func() *retry.RetryError {
			res, listErr := api.ListInstanceLogs(&rdbSDK.ListInstanceLogsRequest{
				Region:     region,
				InstanceID: instanceID,
				OrderBy:    rdbSDK.ListInstanceLogsRequestOrderByCreatedAtDesc,
			}, scw.WithContext(ctx))
			if listErr != nil {
				return retry.NonRetryableError(fmt.Errorf("failed to list instance logs: %w", listErr))
			}

			if len(res.InstanceLogs) > 0 {
				return nil
			}

			return retry.RetryableError(fmt.Errorf("waiting for prepared logs on instance %s", instanceID))
		})
	}
}
