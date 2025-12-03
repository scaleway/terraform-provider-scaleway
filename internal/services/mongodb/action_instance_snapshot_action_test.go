package mongodb_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccActionMongoDBInstanceSnapshot_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccActionMongoDBInstanceSnapshot_Basic because actions are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_mongodb_instance" "main" {
						name        = "test-mongodb-action-snapshot"
						version     = "7.0.12"
						node_type   = "MGDB-PLAY2-NANO"
						node_number = 1
						user_name   = "my_initial_user"
						password    = "thiZ_is_v&ry_s3cret"

						lifecycle {
							action_trigger {
								events  = [after_create]
								actions = [action.scaleway_mongodb_instance_snapshot_action.main]
							}
						}
					}

					action "scaleway_mongodb_instance_snapshot_action" "main" {
						config {
							instance_id = scaleway_mongodb_instance.main.id
							name        = "tf-acc-mongodb-instance-snapshot-action"
							expires_at  = "2026-11-01T00:00:00Z"
							wait        = true
						}
					}
				`,
			},
		},
	})
}
