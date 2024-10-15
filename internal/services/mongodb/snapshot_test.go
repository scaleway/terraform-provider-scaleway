package mongodb_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	mongodbSDK "github.com/scaleway/scaleway-sdk-go/api/mongodb/v1alpha1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/mongodb"
)

func TestAccMongoDBSnapshot_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      isSnapshotDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_mongodb_instance" "main" {
						name       = "test-mongodb-instance"
						version    = "7.0.12"
						node_type  = "MGDB-PRO2-XXS"
						node_number = 1
						user_name  = "my_initial_user"
						password   = "thiZ_is_v&ry_s3cret"
					}

					resource "scaleway_mongodb_snapshot" "main" {
						instance_id = scaleway_mongodb_instance.main.id
						name        = "test-snapshot"
						expires_at  = "2024-12-31T23:59:59Z"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_mongodb_snapshot.main", "name", "test-snapshot"),
				),
			},
		},
	})
}

func TestAccMongoDBSnapshot_Update(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      isSnapshotDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_mongodb_instance" "main" {
						name       = "test-mongodb-instance"
						version    = "7.0.12"
						node_type  = "MGDB-PLAY2-NANO"
						node_number = 1
						user_name  = "my_initial_user"
						password   = "thiZ_is_v&ry_s3cret"
					}

					resource "scaleway_mongodb_snapshot" "main" {
						instance_id = scaleway_mongodb_instance.main.id
						name        = "test-snapshot"
						expires_at  = "2024-12-31T23:59:59Z"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_mongodb_snapshot.main", "expires_at", "2024-12-31T23:59:59Z"),
				),
			},
			{
				Config: `
					resource "scaleway_mongodb_instance" "main" {
						name       = "test-mongodb-instance"
						version    = "7.0.12"
						node_type  = "MGDB-PLAY2-NANO"
						node_number = 1
						user_name  = "my_initial_user"
						password   = "thiZ_is_v&ry_s3cret"
					}

					resource "scaleway_mongodb_snapshot" "main" {
						instance_id = scaleway_mongodb_instance.main.id
						name        = "updated-snapshot"
						expires_at  = "2025-09-20T23:59:59Z"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_mongodb_snapshot.main", "name", "updated-snapshot"),
					resource.TestCheckResourceAttr("scaleway_mongodb_snapshot.main", "expires_at", "2025-09-20T23:59:59Z"),
				),
			},
		},
	})
}

func isSnapshotDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_mongodb_snapshot" {
				continue
			}

			mongodbAPI, region, ID, err := mongodb.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}
			instanceID := zonal.ExpandID(regional.ExpandID(rs.Primary.Attributes["instance_id"]).String())

			listSnapshots, err := mongodbAPI.ListSnapshots(&mongodbSDK.ListSnapshotsRequest{
				InstanceID: &instanceID.ID,
				Region:     region,
			})
			if err != nil {
				return err
			}

			for _, snapshot := range listSnapshots.Snapshots {
				if snapshot.ID == ID {
					return fmt.Errorf("snapshot (%s) still exists", rs.Primary.ID)
				}
			}
		}
		return nil
	}
}
