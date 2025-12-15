package rdb_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccActionRDBReadReplicaPromote_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccActionRDBReadReplicaPromote_Basic because action are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create instance + read replica, action promotes it
			{
				Config: `
					resource "scaleway_rdb_instance" "main" {
						name = "test-rdb-action-read-replica-promote"
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
								actions = [action.scaleway_rdb_read_replica_promote_action.main]
							}
						}
					}

					action "scaleway_rdb_read_replica_promote_action" "main" {
						config {
							read_replica_id = scaleway_rdb_read_replica.main.id
							wait = true
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("scaleway_rdb_instance.main", "id"),
				),
				// ExpectNonEmptyPlan is required because the promote action destroys the read replica,
				// causing Terraform to detect drift and plan to recreate it.
				ExpectNonEmptyPlan: true,
			},
			// Step 2: Read replica has been promoted and removed from config
			{
				Config: `
					resource "scaleway_rdb_instance" "main" {
						name = "test-rdb-action-read-replica-promote"
						node_type = "db-dev-s"
						engine = "PostgreSQL-15"
						is_ha_cluster = false
						disable_backup = true
						user_name = "my_initial_user"
						password = "thiZ_is_v&ry_s3cret"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_rdb_instance.main", "name", "test-rdb-action-read-replica-promote"),
					resource.TestCheckResourceAttr("scaleway_rdb_instance.main", "engine", "PostgreSQL-15"),
				),
			},
		},
	})
}
