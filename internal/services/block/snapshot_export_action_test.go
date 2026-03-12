package block_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	blocktestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/block/testfuncs"
	objectchecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/object/testfuncs"
)

func TestAccActionSnapshotExport_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccActionSnapshotExport_Basic because action are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	bucketName := sdkacctest.RandomWithPrefix("test-acc-scaleway-export-snapshot")
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			blocktestfuncs.IsSnapshotDestroyed(tt),
			objectchecks.IsObjectDestroyed(tt),
			objectchecks.IsBucketDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_block_volume" "main" {
						iops = 5000
						size_in_gb = 10
					}

					resource "scaleway_block_snapshot" "main" {
						name = "test-terraform-export-snapshot-basic"
						volume_id = scaleway_block_volume.main.id

  						lifecycle {
      						action_trigger {
          						events  = [after_create]
          						actions = [action.scaleway_block_export_snapshot.main]
      						}
  						}
					}

					resource "scaleway_object_bucket" "export-bucket" {
						name = "%s"
						force_destroy = true
					}

					action "scaleway_block_export_snapshot" "main" {
						config {
							snapshot_id = scaleway_block_snapshot.main.id
							bucket = scaleway_object_bucket.export-bucket.name
							key = "exported-snapshot.qcow2"
							wait = true
						}
					}
				`, bucketName),
				Check: resource.ComposeTestCheckFunc(
					blocktestfuncs.IsSnapshotPresent(tt, "scaleway_block_snapshot.main"),
					acctest.CheckResourceAttrUUID("scaleway_block_snapshot.main", "id"),
					resource.TestCheckResourceAttr("scaleway_block_snapshot.main", "name", "test-terraform-export-snapshot-basic"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_block_volume" "main" {
						iops = 5000
						size_in_gb = 10
					}

					resource "scaleway_block_snapshot" "main" {
						name = "test-terraform-export-snapshot-basic"
						volume_id = scaleway_block_volume.main.id

  						lifecycle {
      						action_trigger {
          						events  = [after_create]
          						actions = [action.scaleway_block_export_snapshot.main]
      						}
  						}
					}

					resource "scaleway_object_bucket" "export-bucket" {
						name = "%s"
						force_destroy = true
					}

					action "scaleway_block_export_snapshot" "main" {
						config {
							snapshot_id = scaleway_block_snapshot.main.id
							bucket = scaleway_object_bucket.export-bucket.name
							key = "exported-snapshot.qcow2"
							wait = true
						}
					}

					data "scaleway_object" "main" {
  						bucket = scaleway_object_bucket.export-bucket.name
  						key    = "exported-snapshot.qcow2"
					}
				`, bucketName),
				Check: resource.ComposeTestCheckFunc(
					blocktestfuncs.IsSnapshotPresent(tt, "scaleway_block_snapshot.main"),
					acctest.CheckResourceAttrUUID("scaleway_block_snapshot.main", "id"),
					resource.TestCheckResourceAttr("scaleway_block_snapshot.main", "name", "test-terraform-export-snapshot-basic"),
					objectchecks.IsObjectExists(tt, "data.scaleway_object.main"),
				),
			},
		},
	})
}
