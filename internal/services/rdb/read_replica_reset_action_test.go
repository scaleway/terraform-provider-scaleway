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

func TestAccActionRDBReadReplicaReset_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccActionRDBReadReplicaReset_Basic because action are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_rdb_instance" "main" {
						name = "test-rdb-action-read-replica-reset"
						node_type = "db-dev-s"
						engine = "PostgreSQL-15"
						is_ha_cluster = false
						disable_backup = true
						user_name = "my_initial_user"
						password = "thiZ_is_v&ry_s3cret"
					}

					resource "scaleway_rdb_read_replica" "main" {
						instance_id = scaleway_rdb_instance.main.id
						direct_access {}

						lifecycle {
							action_trigger {
								events  = [after_create]
								actions = [action.scaleway_rdb_read_replica_reset.main]
							}
						}
					}

					action "scaleway_rdb_read_replica_reset" "main" {
						config {
							read_replica_id = scaleway_rdb_read_replica.main.id
							wait = true
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isReadReplicaReset(tt, "scaleway_rdb_read_replica.main"),
				),
			},
		},
	})
}

func isReadReplicaReset(tt *acctest.TestTools, readReplicaResourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[readReplicaResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", readReplicaResourceName)
		}

		region, id, err := regional.ParseID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("failed to parse read replica ID: %w", err)
		}

		api := rdbSDK.NewAPI(tt.Meta.ScwClient())

		readReplica, err := api.GetReadReplica(&rdbSDK.GetReadReplicaRequest{
			Region:        region,
			ReadReplicaID: id,
		}, scw.WithContext(context.Background()))
		if err != nil {
			return fmt.Errorf("failed to get read replica: %w", err)
		}

		if readReplica == nil {
			return fmt.Errorf("read replica %s not found", id)
		}

		if readReplica.Status != rdbSDK.ReadReplicaStatusReady && readReplica.Status != rdbSDK.ReadReplicaStatusProvisioning {
			return fmt.Errorf("read replica %s is not in a valid state, status: %s", id, readReplica.Status)
		}

		return nil
	}
}
