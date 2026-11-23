package mongodb_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	mongodbSDK "github.com/scaleway/scaleway-sdk-go/api/mongodb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
)

func TestAccActionMongoDBInstanceSnapshot_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccActionMongoDBInstanceSnapshot_Basic because actions are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	expiresAt := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_mongodb_instance" "main" {
						name        = "test-mongodb-action-snapshot"
						version     = "7.0"
						node_type   = "MGDB-PLAY2-NANO"
						node_number = 1
						user_name   = "my_initial_user"
						password    = "thiZ_is_v&ry_s3cret"

						lifecycle {
							action_trigger {
								events  = [after_create]
								actions = [action.scaleway_mongodb_instance_snapshot.main]
							}
						}
					}

					action "scaleway_mongodb_instance_snapshot" "main" {
						config {
							instance_id = scaleway_mongodb_instance.main.id
							name        = "tf-acc-mongodb-instance-snapshot-action"
							expires_at  = %q
							wait        = true
						}
					}
				`, expiresAt),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_mongodb_instance" "main" {
						name        = "test-mongodb-action-snapshot"
						version     = "7.0"
						node_type   = "MGDB-PLAY2-NANO"
						node_number = 1
						user_name   = "my_initial_user"
						password    = "thiZ_is_v&ry_s3cret"

						lifecycle {
							action_trigger {
								events  = [after_create]
								actions = [action.scaleway_mongodb_instance_snapshot.main]
							}
						}
					}

					action "scaleway_mongodb_instance_snapshot" "main" {
						config {
							instance_id = scaleway_mongodb_instance.main.id
							name        = "tf-acc-mongodb-instance-snapshot-action"
							expires_at  = %q
							wait        = true
						}
					}
				`, expiresAt),
				Check: resource.ComposeTestCheckFunc(
					isSnapshotCreated(tt, "scaleway_mongodb_instance.main", "tf-acc-mongodb-instance-snapshot-action"),
				),
			},
		},
	})
}

func isSnapshotCreated(tt *acctest.TestTools, instanceResourceName, snapshotName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[instanceResourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", instanceResourceName)
		}

		instanceID := rs.Primary.ID

		region, id, err := regional.ParseID(instanceID)
		if err != nil {
			return fmt.Errorf("failed to parse instance ID: %w", err)
		}

		api := mongodbSDK.NewAPI(tt.Meta.ScwClient())

		snapshots, err := api.ListSnapshots(&mongodbSDK.ListSnapshotsRequest{
			Region:     region,
			InstanceID: &id,
		}, scw.WithAllPages(), scw.WithContext(context.Background()))
		if err != nil {
			return fmt.Errorf("failed to list snapshots: %w", err)
		}

		for _, snapshot := range snapshots.Snapshots {
			if snapshot.Name == snapshotName {
				return nil
			}
		}

		return fmt.Errorf("snapshot with name %q not found for instance %s", snapshotName, instanceID)
	}
}
