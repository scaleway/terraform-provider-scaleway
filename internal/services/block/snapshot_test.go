package block_test

import (
	"fmt"
	"regexp"
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
		ProtoV6ProviderFactories: tt.ProviderFactories,
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
					resource.TestMatchResourceAttr("scaleway_block_snapshot.main", "srn", regexp.MustCompile(`^srn://block\..+/zones/.+/snapshots/.+$`)),
				),
			},
		},
	})
}

// TestAccSnapshot_RefreshBareUUIDIdentity reproduces the v2.80.0 -> v2.88.0
// upgrade regression where a refresh failed with:
//
//	Failed to parse block snapshot ID
//	cant parse localized id: <bare uuid>
//
// The zonal identity schema stores the snapshot id as a bare UUID. v2.88.0 fed
// that bare UUID straight into zonal.ParseID (which expects "zone/uuid"), so
// every refresh after the SDKv2 -> Framework migration blew up. The second
// step forces a refresh of the state written by step 1, exercising the Read
// path with a bare-UUID identity.
func TestAccSnapshot_RefreshBareUUIDIdentity(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             blocktestfuncs.IsSnapshotDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_block_volume main {
						iops = 5000
						size_in_gb = 10
					}

					resource scaleway_block_snapshot main {
						name = "test-block-snapshot-refresh"
						volume_id = scaleway_block_volume.main.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					blocktestfuncs.IsSnapshotPresent(tt, "scaleway_block_snapshot.main"),
					acctest.CheckResourceAttrUUID("scaleway_block_snapshot.main", "id"),
				),
			},
			{
				// Same config: Terraform refreshes the existing resource before
				// planning. This is the step that errored in v2.88.0 when Read
				// parsed the bare-UUID identity id as a localized id.
				Config: `
					resource scaleway_block_volume main {
						iops = 5000
						size_in_gb = 10
					}

					resource scaleway_block_snapshot main {
						name = "test-block-snapshot-refresh"
						volume_id = scaleway_block_volume.main.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					blocktestfuncs.IsSnapshotPresent(tt, "scaleway_block_snapshot.main"),
					acctest.CheckResourceAttrUUID("scaleway_block_snapshot.main", "id"),
				),
			},
		},
	})
}

func TestAccSnapshot_VolumeIDConflict(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             blocktestfuncs.IsSnapshotDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_block_volume main {
						iops = 5000
						size_in_gb = 10
					}

					resource scaleway_block_snapshot main {
						name = "test-block-snapshot-conflict"
						volume_id = scaleway_block_volume.main.id
						import {
							bucket = "tf-test-block-snapshot-conflict"
							key    = "test-acc-block-snapshot-conflict.qcow2"
						}
					}
				`,
				ExpectError: regexp.MustCompile(`"volume_id" cannot be specified when "import" is specified`),
			},
		},
	})
}

func TestAccSnapshot_FromS3(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	bucketName := sdkacctest.RandomWithPrefix("tf-test-scaleway-block-snapshot")
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

	bucketName := sdkacctest.RandomWithPrefix("tf-test-scaleway-export-block-snapshot")
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
