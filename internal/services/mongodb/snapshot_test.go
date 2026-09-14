package mongodb_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	mongodbSDK "github.com/scaleway/scaleway-sdk-go/api/mongodb/v1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/mongodb"
)

func TestAccMongoDBSnapshot_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	expiresAt := time.Now().AddDate(0, 6, 0).UTC().Format(time.RFC3339)
	if !*acctest.UpdateCassettes {
		// Hardcoded value must match expires_at in the cassette request body.
		expiresAt = "2027-03-10T23:59:59Z"
	}

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isSnapshotDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
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
						expires_at  = %q
					}
				`, expiresAt),
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

	expiresAt := time.Now().AddDate(0, 6, 0).UTC().Format(time.RFC3339)
	if !*acctest.UpdateCassettes {
		// Hardcoded value must match expires_at in the cassette request body.
		expiresAt = "2027-03-10T23:59:59Z"
	}

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isSnapshotDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
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
						expires_at  = %q
					}
				`, expiresAt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_mongodb_snapshot.main", "expires_at", expiresAt),
				),
			},
			{
				Config: fmt.Sprintf(`
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
						expires_at  = %q
					}
				`, expiresAt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_mongodb_snapshot.main", "name", "updated-snapshot"),
					resource.TestCheckResourceAttr("scaleway_mongodb_snapshot.main", "expires_at", expiresAt),
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

			_, err = mongodbAPI.GetSnapshot(&mongodbSDK.GetSnapshotRequest{
				SnapshotID: ID,
				Region:     region,
			})
			if err == nil {
				return fmt.Errorf("instance (%s) still exists", rs.Primary.ID)
			}

			if !httperrors.Is404(err) {
				return err
			}
		}

		return nil
	}
}
