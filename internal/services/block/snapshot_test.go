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

func TestAccSnapshot_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             blocktestfuncs.IsSnapshotDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_block_volume main {
						iops = 5000
						size_in_gb = 10
					}

					resource scaleway_block_snapshot main {
						name = "test-block-snapshot-basic"
						volume_id = scaleway_block_volume.main.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					blocktestfuncs.IsSnapshotPresent(tt, "scaleway_block_snapshot.main"),
					acctest.CheckResourceAttrUUID("scaleway_block_snapshot.main", "id"),
					resource.TestCheckResourceAttr("scaleway_block_snapshot.main", "name", "test-block-snapshot-basic"),
				),
			},
		},
	})
}

func TestAccSnapshot_FromS3(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	bucketName := sdkacctest.RandomWithPrefix("test-acc-scaleway-block-snapshot")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			blocktestfuncs.IsSnapshotDestroyed(tt),
			objectchecks.IsObjectDestroyed(tt),
			objectchecks.IsBucketDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "snapshot-bucket" {
					  name = "%s"
					}
					
					resource "scaleway_object" "qcow-object" {
					  bucket = scaleway_object_bucket.snapshot-bucket.name
					  key    = "test-acc-block-snapshot.qcow2"
					  file   = "testfixture/small_image.qcow2"
					}

					resource "scaleway_block_snapshot" "qcow-block-snapshot" {
					  name = "test-acc-block-snapshot-qcow2"
					  import {
					    bucket = scaleway_object.qcow-object.bucket
					    key    = scaleway_object.qcow-object.key
					  }
					}
				`, bucketName),
				Check: resource.ComposeTestCheckFunc(
					blocktestfuncs.IsSnapshotPresent(tt, "scaleway_block_snapshot.qcow-block-snapshot"),
					acctest.CheckResourceAttrUUID("scaleway_block_snapshot.qcow-block-snapshot", "id"),
					resource.TestCheckResourceAttr("scaleway_block_snapshot.qcow-block-snapshot", "name", "test-acc-block-snapshot-qcow2"),
				),
			},
		},
	})
}

func TestAccSnapshot_ToS3(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	bucketName := sdkacctest.RandomWithPrefix("test-acc-scaleway-export-block-snapshot")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			blocktestfuncs.IsSnapshotDestroyed(tt),
			objectchecks.IsObjectDestroyed(tt),
			objectchecks.IsBucketDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource scaleway_block_volume main {
						iops = 5000
						size_in_gb = 10
					}

					resource "scaleway_object_bucket" "snapshot-bucket" {
					  name = "%s"
					}
					
					resource "scaleway_object" "qcow-object" {
					  bucket = scaleway_object_bucket.snapshot-bucket.name
					  key    = "test-acc-export-block-snapshot-qcow2"
					  content = "test"
					}

					resource "scaleway_block_snapshot" "qcow-block-snapshot" {
					  name = "test-acc-export-block-snapshot-qcow2"
					  volume_id = scaleway_block_volume.main.id
					  export {
					    bucket = scaleway_object.qcow-object.bucket
					    key    = scaleway_object.qcow-object.key
					  }
					}
				`, bucketName),
				Check: resource.ComposeTestCheckFunc(
					blocktestfuncs.IsSnapshotPresent(tt, "scaleway_block_snapshot.qcow-block-snapshot"),
					acctest.CheckResourceAttrUUID("scaleway_block_snapshot.qcow-block-snapshot", "id"),
					resource.TestCheckResourceAttr("scaleway_block_snapshot.qcow-block-snapshot", "name", "test-acc-export-block-snapshot-qcow2"),
					objectchecks.IsObjectExists(tt, "scaleway_object.qcow-object"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource scaleway_block_volume main {
						iops = 5000
						size_in_gb = 10
					}

					resource "scaleway_object_bucket" "snapshot-bucket" {
					  name = "%s"
					}
					
					resource "scaleway_object" "qcow-object" {
					  bucket = scaleway_object_bucket.snapshot-bucket.name
					  key    = "test-acc-export-block-snapshot-qcow2"
					  content = "test"
					}

					resource "scaleway_block_snapshot" "qcow-block-export-snapshot" {
					  name = "test-acc-export-block-snapshot-qcow2"
					  volume_id = scaleway_block_volume.main.id
					  export {
					    bucket = scaleway_object.qcow-object.bucket
					    key    = scaleway_object.qcow-object.key
					  }
					}

					resource "scaleway_block_snapshot" "qcow-block-import-snapshot" {
					  name = "test-acc-block-snapshot-qcow2"
					  import {
					    bucket = scaleway_object.qcow-object.bucket
					    key    = scaleway_object.qcow-object.key
					  }
					}

					resource scaleway_block_volume new-volume {
						name = "test-block-volume-from-snapshot"
						iops = 5000
						snapshot_id = scaleway_block_snapshot.qcow-block-import-snapshot.id
					}

				`, bucketName),
				Check: resource.ComposeTestCheckFunc(
					blocktestfuncs.IsSnapshotPresent(tt, "scaleway_block_snapshot.qcow-block-export-snapshot"),
					acctest.CheckResourceAttrUUID("scaleway_block_snapshot.qcow-block-export-snapshot", "id"),
					resource.TestCheckResourceAttr("scaleway_block_snapshot.qcow-block-export-snapshot", "name", "test-acc-export-block-snapshot-qcow2"),
					objectchecks.IsObjectExists(tt, "scaleway_object.qcow-object"),
					blocktestfuncs.IsSnapshotPresent(tt, "scaleway_block_snapshot.qcow-block-import-snapshot"),
					acctest.CheckResourceAttrUUID("scaleway_block_snapshot.qcow-block-import-snapshot", "id"),
					resource.TestCheckResourceAttr("scaleway_block_volume.new-volume", "size_in_gb", "10"),
					resource.TestCheckResourceAttr("scaleway_block_volume.new-volume", "iops", "5000"),
					resource.TestCheckResourceAttrPair("scaleway_block_volume.new-volume", "snapshot_id", "scaleway_block_snapshot.qcow-block-import-snapshot", "id"),
				),
			},
		},
	})
}
