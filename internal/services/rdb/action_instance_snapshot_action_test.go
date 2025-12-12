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
)

func TestAccActionRDBInstanceSnapshot_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccActionRDBInstanceSnapshot_Basic because action are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_rdb_instance" "main" {
						name = "test-rdb-action-snapshot"
						node_type = "db-dev-s"
						engine = "PostgreSQL-15"
						is_ha_cluster = false
						disable_backup = true
						user_name = "my_initial_user"
						password = "thiZ_is_v&ry_s3cret"
						volume_type = "sbs_5k"
						volume_size_in_gb = 10

						lifecycle {
							action_trigger {
								events  = [after_create]
								actions = [action.scaleway_rdb_instance_snapshot_action.main]
							}
						}
					}

					action "scaleway_rdb_instance_snapshot_action" "main" {
						config {
							instance_id = scaleway_rdb_instance.main.id
							name = "tf-rdb-snapshot"
							wait = true
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isSnapshotCreated(tt, "scaleway_rdb_instance.main", "tf-rdb-snapshot"),
				),
			},
		},
	})
}

func isSnapshotCreated(tt *acctest.TestTools, instanceResourceName, snapshotName string) resource.TestCheckFunc {
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

		snapshots, err := api.ListSnapshots(&rdbSDK.ListSnapshotsRequest{
			Region:     region,
			InstanceID: scw.StringPtr(id),
			Name:       scw.StringPtr(snapshotName),
		}, scw.WithContext(context.Background()))
		if err != nil {
			return fmt.Errorf("failed to list snapshots: %w", err)
		}

		if len(snapshots.Snapshots) == 0 {
			return fmt.Errorf("snapshot %s not found for instance %s", snapshotName, id)
		}

		snapshot := snapshots.Snapshots[0]
		if snapshot.Status != rdbSDK.SnapshotStatusReady && snapshot.Status != rdbSDK.SnapshotStatusCreating {
			return fmt.Errorf("snapshot %s is not in a valid state, status: %s", snapshotName, snapshot.Status)
		}

		return nil
	}
}
