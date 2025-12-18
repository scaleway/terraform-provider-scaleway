package instance_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	instancechecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/instance/testfuncs"
	objectchecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/object/testfuncs"
)

func TestAccSnapshot_Server(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             instancechecks.IsVolumeDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_server" "main" {
						image = "ubuntu_focal"
						type = "DEV1-S"
						root_volume {
							volume_type = "l_ssd"
						}
					}

					resource "scaleway_instance_snapshot" "main" {
						volume_id = scaleway_instance_server.main.root_volume.0.volume_id
					}`,
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsSnapshotPresent(tt, "scaleway_instance_snapshot.main"),
				),
			},
		},
	})
}

func TestAccSnapshot_FromS3(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	bucketName := sdkacctest.RandomWithPrefix("test-acc-scaleway-instance-snapshot")
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			instancechecks.IsSnapshotDestroyed(tt),
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
					  key    = "test-acc-instance-snapshot.qcow2"
					  file   = "testfixture/small_image.qcow2"
					}

					resource "scaleway_instance_snapshot" "qcow-instance-snapshot" {
					  name = "test-acc-snapshot-import-default"
					  import {
					    bucket = scaleway_object.qcow-object.bucket
					    key    = scaleway_object.qcow-object.key
					  }
					}
				`, bucketName),
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsSnapshotPresent(tt, "scaleway_instance_snapshot.qcow-instance-snapshot"),
					acctest.CheckResourceAttrUUID("scaleway_instance_snapshot.qcow-instance-snapshot", "id"),
					resource.TestCheckResourceAttr("scaleway_instance_snapshot.qcow-instance-snapshot", "name", "test-acc-snapshot-import-default"),
					resource.TestCheckResourceAttr("scaleway_instance_snapshot.qcow-instance-snapshot", "type", "l_ssd"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "snapshot-bucket" {
					  name = "%s"
					}
					
					resource "scaleway_object" "qcow-object" {
					  bucket = scaleway_object_bucket.snapshot-bucket.name
					  key    = "test-acc-instance-snapshot.qcow2"
					  file   = "testfixture/small_image.qcow2"
					}

					resource "scaleway_instance_snapshot" "qcow-instance-snapshot" {
					  name = "test-acc-snapshot-import-lssd"
					  type = "l_ssd"
					  import {
					    bucket = scaleway_object.qcow-object.bucket
					    key    = scaleway_object.qcow-object.key
					  }
					}
				`, bucketName),
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsSnapshotPresent(tt, "scaleway_instance_snapshot.qcow-instance-snapshot"),
					acctest.CheckResourceAttrUUID("scaleway_instance_snapshot.qcow-instance-snapshot", "id"),
					resource.TestCheckResourceAttr("scaleway_instance_snapshot.qcow-instance-snapshot", "name", "test-acc-snapshot-import-lssd"),
					resource.TestCheckResourceAttr("scaleway_instance_snapshot.qcow-instance-snapshot", "type", "l_ssd"),
				),
			},
		},
	})
}
