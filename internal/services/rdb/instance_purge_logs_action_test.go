package rdb_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	rdbSDK "github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

func TestAccActionRDBInstanceLogsPurge_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccActionRDBInstanceLogsPurge_Basic because action are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_rdb_instance" "main" {
						name = "test-rdb-action-logs-purge"
						node_type = "db-dev-s"
						engine = "PostgreSQL-15"
						is_ha_cluster = false
						disable_backup = true
						user_name = "my_initial_user"
						password = "thiZ_is_v&ry_s3cret"

						lifecycle {
							action_trigger {
								events  = [after_create]
								actions = [action.scaleway_rdb_instance_purge_logs.main]
							}
						}
					}

					action "scaleway_rdb_instance_purge_logs" "main" {
						config {
							instance_id = scaleway_rdb_instance.main.id
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isInstanceLogsPurged(tt, "scaleway_rdb_instance.main"),
				),
			},
		},
	})
}

func isInstanceLogsPurged(tt *acctest.TestTools, instanceResourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[instanceResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", instanceResourceName)
		}

		region, id, err := regional.ParseID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("failed to parse instance ID: %w", err)
		}

		api := rdbSDK.NewAPI(tt.Meta.ScwClient())

		var instance *rdbSDK.Instance

		err = transport.RetryOn403(context.Background(), func() error {
			var err error

			instance, err = api.GetInstance(&rdbSDK.GetInstanceRequest{
				Region:     region,
				InstanceID: id,
			}, scw.WithContext(context.Background()))

			return err
		})
		if err != nil {
			return fmt.Errorf("failed to get instance: %w", err)
		}

		if instance == nil {
			return fmt.Errorf("instance %s not found", id)
		}

		if instance.Status != rdbSDK.InstanceStatusReady {
			return fmt.Errorf("instance %s is not ready, status: %s", id, instance.Status)
		}

		return nil
	}
}
