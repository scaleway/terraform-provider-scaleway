package instance_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	instanceSDK "github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	blocktestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/block/testfuncs"
	instancechecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/instance/testfuncs"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/object"
	objectchecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/object/testfuncs"
)

const snapshotKey = "exported-snapshot.qcow2"

func TestAccAction_InstanceExportSnapshot_Local(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccAction_InstanceExportSnapshot_Local because action are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	var bucketName string
	for {
		bucketName = sdkacctest.RandomWithPrefix("test-acc-action-instance-export-snap-local")
		if len(bucketName) < 63 {
			break
		}
	}

	size := 10

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			instancechecks.IsServerDestroyed(tt),
			instancechecks.IsVolumeDestroyed(tt),
			instancechecks.IsSnapshotDestroyed(tt),
			objectchecks.IsObjectDestroyed(tt),
			objectchecks.IsBucketDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_instance_server" "to_snapshot" {
						name = "test-tf-action-instance-export-snap-local"
						type = "DEV1-S"
						image = "ubuntu_jammy"

						root_volume {
							volume_type = "l_ssd"
							size_in_gb = %d
						}
					}

					resource "scaleway_instance_snapshot" "main" {
						volume_id = scaleway_instance_server.to_snapshot.root_volume.0.volume_id

						lifecycle {
							action_trigger {
						  		events  = [after_create]
						  		actions = [action.scaleway_instance_export_snapshot.main]
							}
					  	}
					}

					resource "scaleway_object_bucket" "main" {
					    name = "%s"
					}

					action "scaleway_instance_export_snapshot" "main" {
						config {
						  	snapshot_id = scaleway_instance_snapshot.main.id
							bucket = scaleway_object_bucket.main.name
							key = %q
						}
					}`, size, bucketName, snapshotKey),
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsSnapshotPresent(tt, "scaleway_instance_snapshot.main"),
					objectchecks.CheckBucketExists(tt, "scaleway_object_bucket.main", true),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_instance_server" "to_snapshot" {
						name = "test-tf-action-instance-export-snap-local"
						type = "DEV1-S"
						image = "ubuntu_jammy"
			
						root_volume {
							volume_type = "l_ssd"
							size_in_gb = %d
						}
					}
			
					resource "scaleway_instance_snapshot" "main" {
						volume_id = scaleway_instance_server.to_snapshot.root_volume.0.volume_id
			
						lifecycle {
							action_trigger {
						  		events  = [after_create]
						  		actions = [action.scaleway_instance_export_snapshot.main]
							}
					  	}
					}
			
					resource "scaleway_object_bucket" "main" {
					    name = "%s"
					}
			
					data "scaleway_object" "qcow-object" {
						bucket = scaleway_object_bucket.main.name
						key = %[3]q
					}
			
					action "scaleway_instance_export_snapshot" "main" {
						config {
						  	snapshot_id = scaleway_instance_snapshot.main.id
							bucket = scaleway_object_bucket.main.name
							key =  %[3]q
						}
					}`, size, bucketName, snapshotKey),
				// The check on this step is the data source.
				// If the snapshot does not exist, the data source will raise an error
			},
			{
				Config: fmt.Sprintf(`			
					resource "scaleway_object_bucket" "main" {
					    name = "%s"
					}`, bucketName),
				Check: destroyUntrackedObjects(t.Context(), tt.Meta, bucketName, "fr-par"),
			},
		},
	})
}

func TestAccAction_InstanceExportSnapshot_SBS(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccAction_InstanceExportSnapshot_SBS because action are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	var bucketName string
	for {
		bucketName = sdkacctest.RandomWithPrefix("test-acc-action-instance-export-snap-sbs")
		if len(bucketName) < 63 {
			break
		}
	}

	size := 10

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			instancechecks.IsServerDestroyed(tt),
			blocktestfuncs.IsVolumeDestroyed(tt),
			blocktestfuncs.IsSnapshotDestroyed(tt),
			objectchecks.IsObjectDestroyed(tt),
			objectchecks.IsBucketDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_instance_server" "to_snapshot" {
						name = "test-tf-action-instance-export-snap-sbs"
						type = "DEV1-S"
						image = "ubuntu_jammy"

						root_volume {
							size_in_gb = %d
						}
					}

					resource "scaleway_block_snapshot" "main" {
						volume_id = scaleway_instance_server.to_snapshot.root_volume.0.volume_id

						lifecycle {
							action_trigger {
						  		events  = [after_create]
						  		actions = [action.scaleway_instance_export_snapshot.main]
							}
					  	}
					}

					resource "scaleway_object_bucket" "main" {
					    name = "%s"
					}

					action "scaleway_instance_export_snapshot" "main" {
						config {
						  	snapshot_id = scaleway_block_snapshot.main.id
							bucket = scaleway_object_bucket.main.name
							key = %q
						}
					}`, size, bucketName, snapshotKey),
				Check: resource.ComposeTestCheckFunc(
					blocktestfuncs.IsSnapshotPresent(tt, "scaleway_block_snapshot.main"),
					objectchecks.CheckBucketExists(tt, "scaleway_object_bucket.main", true),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_instance_server" "to_snapshot" {
						name = "test-tf-action-instance-export-snap-sbs"
						type = "DEV1-S"
						image = "ubuntu_jammy"
			
						root_volume {
							size_in_gb = %d
						}
					}
			
					resource "scaleway_block_snapshot" "main" {
						volume_id = scaleway_instance_server.to_snapshot.root_volume.0.volume_id
			
						lifecycle {
							action_trigger {
						  		events  = [after_create]
						  		actions = [action.scaleway_instance_export_snapshot.main]
							}
					  	}
					}
			
					resource "scaleway_object_bucket" "main" {
					    name = "%s"
					}
			
					data "scaleway_object" "qcow-object" {
						bucket = scaleway_object_bucket.main.name
						key = %[3]q
					}
			
					action "scaleway_instance_export_snapshot" "main" {
						config {
						  	snapshot_id = scaleway_block_snapshot.main.id
							bucket = scaleway_object_bucket.main.name
							key =  %[3]q
						}
					}`, size, bucketName, snapshotKey),
				// The check on this step is the data source.
				// If the snapshot does not exist, the data source will raise an error
			},
			{
				Config: fmt.Sprintf(`			
					resource "scaleway_object_bucket" "main" {
					    name = "%s"
					}`, bucketName),
				Check: destroyUntrackedObjects(t.Context(), tt.Meta, bucketName, "fr-par"),
			},
		},
	})
}

func TestAccAction_InstanceExportSnapshot_Wait(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccAction_InstanceExportSnapshot_Wait because action are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	var bucketName string
	for {
		bucketName = sdkacctest.RandomWithPrefix("test-acc-action-instance-export-snap-wait")
		if len(bucketName) < 63 {
			break
		}
	}

	size := 10

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			instancechecks.IsServerDestroyed(tt),
			instancechecks.IsVolumeDestroyed(tt),
			instancechecks.IsSnapshotDestroyed(tt),
			objectchecks.IsBucketDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_instance_server" "to_snapshot" {
						name = "test-tf-action-instance-export-snapshot-wait"
						type = "DEV1-S"
						image = "ubuntu_jammy"

						root_volume {
							volume_type = "l_ssd"
							size_in_gb = %d
						}
					}

					resource "scaleway_instance_snapshot" "main" {
						volume_id = scaleway_instance_server.to_snapshot.root_volume.0.volume_id

						lifecycle {
							action_trigger {
						  		events  = [after_create]
						  		actions = [action.scaleway_instance_export_snapshot.main]
							}
					  	}
					}

					resource "scaleway_object_bucket" "main" {
					    name = "%s"
					}

					action "scaleway_instance_export_snapshot" "main" {
						config {
						  	snapshot_id = scaleway_instance_snapshot.main.id
							bucket = scaleway_object_bucket.main.name
							key = %q
							wait = true
						}
					}`, size, bucketName, snapshotKey),
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsSnapshotPresent(tt, "scaleway_instance_snapshot.main"),
					objectchecks.CheckBucketExists(tt, "scaleway_object_bucket.main", true),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_object_bucket" "main" {
					    name = "%s"
					}

					data "scaleway_object" "qcow-object" {
						bucket = scaleway_object_bucket.main.name
						key = %[2]q
					}

					resource "scaleway_instance_snapshot" "new_from_exported" {
						import {
					    	bucket = scaleway_object_bucket.main.name
					    	key    = %[2]q
					  	}
					}

					resource "scaleway_instance_volume" "new_from_exported" {
						from_snapshot_id = scaleway_instance_snapshot.new_from_exported.id
						type = "l_ssd"
					}

					resource "scaleway_instance_server" "new_from_exported" {
						name = "test-tf-action-instance-export-snapshot-new"
						type = "DEV1-S"

						root_volume {
							volume_id = scaleway_instance_volume.new_from_exported.id
						}
					}`, bucketName, snapshotKey),
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsSnapshotPresent(tt, "scaleway_instance_snapshot.new_from_exported"),
					instancechecks.IsVolumePresent(tt, "scaleway_instance_volume.new_from_exported"),
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.new_from_exported"),
					resource.TestCheckResourceAttr("scaleway_instance_server.new_from_exported", "root_volume.0.size_in_gb", strconv.Itoa(size)),
					readActualServerState(tt, "scaleway_instance_server.new_from_exported", instanceSDK.ServerStateRunning.String()),
				),
			},
			{
				Config: fmt.Sprintf(`			
					resource "scaleway_object_bucket" "main" {
					    name = "%s"
					}`, bucketName),
				Check: destroyUntrackedObjects(t.Context(), tt.Meta, bucketName, "fr-par"),
			},
		},
	})
}

func TestAccAction_InstanceExportSnapshot_Zone(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccAction_InstanceExportSnapshot_Zone because action are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	var bucketName string
	for {
		bucketName = sdkacctest.RandomWithPrefix("test-acc-action-instance-export-snap-zone")
		if len(bucketName) < 63 {
			break
		}
	}

	size := 10

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			instancechecks.IsServerDestroyed(tt),
			instancechecks.IsVolumeDestroyed(tt),
			instancechecks.IsSnapshotDestroyed(tt),
			objectchecks.IsObjectDestroyed(tt),
			objectchecks.IsBucketDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_instance_server" "to_snapshot" {
						name = "test-tf-action-instance-export-snapshot-zone"
						type = "DEV1-S"
						image = "ubuntu_jammy"
						zone = "nl-ams-2"

						root_volume {
							volume_type = "l_ssd"
							size_in_gb = %d
						}
					}

					resource "scaleway_instance_snapshot" "main" {
						volume_id = scaleway_instance_server.to_snapshot.root_volume.0.volume_id
						zone = "nl-ams-2"

						lifecycle {
							action_trigger {
						  		events  = [after_create]
						  		actions = [action.scaleway_instance_export_snapshot.main]
							}
					  	}
					}

					resource "scaleway_object_bucket" "main" {
					    name = "%s"
						region = "nl-ams"
					}

					action "scaleway_instance_export_snapshot" "main" {
						config {
						  	snapshot_id = scaleway_instance_snapshot.main.id
							bucket = scaleway_object_bucket.main.name
							key = %q
						}
					}`, size, bucketName, snapshotKey),
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsSnapshotPresent(tt, "scaleway_instance_snapshot.main"),
					objectchecks.CheckBucketExists(tt, "scaleway_object_bucket.main", true),
					resource.TestCheckResourceAttr("scaleway_instance_server.to_snapshot", "zone", "nl-ams-2"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_instance_server" "to_snapshot" {
						name = "test-tf-action-instance-export-snapshot-zone"
						type = "DEV1-S"
						image = "ubuntu_jammy"
						zone = "nl-ams-2"
			
						root_volume {
							volume_type = "l_ssd"
							size_in_gb = %d
						}
					}
			
					resource "scaleway_instance_snapshot" "main" {
						volume_id = scaleway_instance_server.to_snapshot.root_volume.0.volume_id
						zone = "nl-ams-2"
			
						lifecycle {
							action_trigger {
						  		events  = [after_create]
						  		actions = [action.scaleway_instance_export_snapshot.main]
							}
					  	}
					}
			
					resource "scaleway_object_bucket" "main" {
					    name = "%s"
						region = "nl-ams"
					}
			
					data "scaleway_object" "qcow-object" {
						bucket = scaleway_object_bucket.main.name
						key = %[3]q
						region = "nl-ams"
					}
			
					action "scaleway_instance_export_snapshot" "main" {
						config {
						  	snapshot_id = scaleway_instance_snapshot.main.id
							bucket = scaleway_object_bucket.main.name
							key =  %[3]q
						}
					}`, size, bucketName, snapshotKey),
				// The check on this step is the data source.
				// If the snapshot does not exist, the data source will raise an error
			},
			{
				Config: fmt.Sprintf(`			
					resource "scaleway_object_bucket" "main" {
					    name = "%s"
						region = "nl-ams"
					}`, bucketName),
				Check: destroyUntrackedObjects(t.Context(), tt.Meta, bucketName, "nl-ams"),
			},
		},
	})
}

func destroyUntrackedObjects(ctx context.Context, meta *meta.Meta, bucketName, region string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		s3Client, err := object.NewS3ClientFromMeta(ctx, meta, region)
		if err != nil {
			return err
		}

		objects, err := s3Client.ListObjects(ctx, &s3.ListObjectsInput{
			Bucket: &bucketName,
		})
		if err != nil {
			return err
		}

		for _, objectKey := range objects.Contents {
			_, err = s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
				Bucket: &bucketName,
				Key:    objectKey.Key,
			})
			if err != nil {
				return err
			}
		}

		return nil
	}
}
