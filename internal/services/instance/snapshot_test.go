package instance_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	instanceSDK "github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/instance"
	instancechecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/instance/testfuncs"
	objectchecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/object/testfuncs"
)

func TestAccSnapshot_Server(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: tt.ProviderFactories,
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
					isSnapshotPresent(tt, "scaleway_instance_snapshot.main"),
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
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			isSnapshotDestroyed(tt),
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
					isSnapshotPresent(tt, "scaleway_instance_snapshot.qcow-instance-snapshot"),
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
					isSnapshotPresent(tt, "scaleway_instance_snapshot.qcow-instance-snapshot"),
					acctest.CheckResourceAttrUUID("scaleway_instance_snapshot.qcow-instance-snapshot", "id"),
					resource.TestCheckResourceAttr("scaleway_instance_snapshot.qcow-instance-snapshot", "name", "test-acc-snapshot-import-lssd"),
					resource.TestCheckResourceAttr("scaleway_instance_snapshot.qcow-instance-snapshot", "type", "l_ssd"),
				),
			},
		},
	})
}

func isSnapshotPresent(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		instanceAPI, zone, ID, err := instance.NewAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = instanceAPI.GetSnapshot(&instanceSDK.GetSnapshotRequest{
			Zone:       zone,
			SnapshotID: ID,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func isSnapshotDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_instance_snapshot" {
				continue
			}

			instanceAPI, zone, ID, err := instance.NewAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = instanceAPI.GetSnapshot(&instanceSDK.GetSnapshotRequest{
				SnapshotID: ID,
				Zone:       zone,
			})

			// If no error resource still exist
			if err == nil {
				return fmt.Errorf("snapshot (%s) still exists", rs.Primary.ID)
			}

			// Unexpected api error we return it
			if !httperrors.Is404(err) {
				return err
			}
		}

		return nil
	}
}
