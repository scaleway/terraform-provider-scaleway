package instance_test

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	blockSDK "github.com/scaleway/scaleway-sdk-go/api/block/v1"
	instanceSDK "github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/block"
	blocktestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/block/testfuncs"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/instance"
	instancechecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/instance/testfuncs"
)

type snapshotSpecsCheck struct {
	Name *string
	Size *scw.Size
	Type *instanceSDK.VolumeVolumeType
	Tags []string
}

func TestAccAction_InstanceCreateSnapshot_Local(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccAction_InstanceCreateSnapshot_Local because action are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	localVolumeType := instanceSDK.VolumeVolumeTypeLSSD

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             instancechecks.IsServerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_instance_server" "main" {
						name = "test-tf-action-instance-create-snapshot-local"
						type = "DEV1-S"
						image = "ubuntu_jammy"

						root_volume {
							volume_type = "%s"
							size_in_gb = 20
						}

						lifecycle {
							action_trigger {
						  		events  = [after_create]
						  		actions = [action.scaleway_instance_create_snapshot.main]
							}
					  	}
					}

					data "scaleway_instance_volume" "main" {
						volume_id = scaleway_instance_server.main.root_volume.0.volume_id
					}

					action "scaleway_instance_create_snapshot" "main" {
						config {
						  	volume_id = scaleway_instance_server.main.root_volume.0.volume_id
							wait = true
						}
					}`, localVolumeType),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsVolumePresent(tt, "data.scaleway_instance_volume.main"),
					checkInstanceSnapshot(tt, "data.scaleway_instance_volume.main", snapshotSpecsCheck{
						Size: scw.SizePtr(20 * scw.GB),
						Type: &localVolumeType,
					}),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_instance_server" "main" {
						name = "test-tf-action-instance-create-snapshot-local"
						type = "DEV1-S"
						image = "ubuntu_jammy"
						tags = [ "add", "tags", "to", "trigger", "update" ]
			
						root_volume {
							volume_type = "%s"
							size_in_gb = 20
						}
			
						lifecycle {
							action_trigger {
						  		events  = [after_update]
						  		actions = [action.scaleway_instance_create_snapshot.main]
							}
					  	}
					}

					data "scaleway_instance_volume" "main" {
						volume_id = scaleway_instance_server.main.root_volume.0.volume_id
					}

					action "scaleway_instance_create_snapshot" "main" {
						config {
						  	volume_id = scaleway_instance_server.main.root_volume.0.volume_id
						  	tags = scaleway_instance_server.main.tags
						  	name = "custom-name-for-snapshot"
							wait = true
						}
					}`, localVolumeType),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsVolumePresent(tt, "data.scaleway_instance_volume.main"),
					checkInstanceSnapshot(tt, "data.scaleway_instance_volume.main", snapshotSpecsCheck{
						Name: scw.StringPtr("custom-name-for-snapshot"),
						Size: scw.SizePtr(20 * scw.GB),
						Tags: []string{"add", "tags", "to", "trigger", "update"},
						Type: &localVolumeType,
					}),
				),
			},
			{
				// This step ensures that the snapshots are deleted before the volume they are based on,
				// otherwise they won't be returned by the ListSnapshot request.
				Config: fmt.Sprintf(`
					resource "scaleway_instance_server" "main" {
						name = "test-tf-action-instance-create-snapshot-local"
						type = "DEV1-S"
						image = "ubuntu_jammy"
						tags = [ "add", "tags", "to", "trigger", "update" ]
			
						root_volume {
							volume_type = "%s"
							size_in_gb = 20
						}
					}

					data "scaleway_instance_volume" "main" {
						volume_id = scaleway_instance_server.main.root_volume.0.volume_id
					}`, localVolumeType),
				Check: destroyUntrackedInstanceSnapshots(tt, "data.scaleway_instance_volume.main"),
			},
		},
	})
}

func TestAccAction_InstanceCreateSnapshot_SBS(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccAction_InstanceCreateSnapshot_SBS because action are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	sbsVolumeType := instanceSDK.VolumeVolumeTypeSbsVolume

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             instancechecks.IsServerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_instance_server" "main" {
						name = "test-tf-action-instance-create-snapshot-sbs"
						type = "DEV1-S"
						image = "ubuntu_jammy"

						root_volume {
							volume_type = "%s"
							size_in_gb = 20
						}

						lifecycle {
							action_trigger {
						  		events  = [after_create]
						  		actions = [action.scaleway_instance_create_snapshot.main]
							}
					  	}
					}

					data "scaleway_block_volume" "main" {
						volume_id = scaleway_instance_server.main.root_volume.0.volume_id
					}

					action "scaleway_instance_create_snapshot" "main" {
						config {
						  	volume_id = scaleway_instance_server.main.root_volume.0.volume_id
							wait = true
						}
					}`, sbsVolumeType),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					blocktestfuncs.IsVolumePresent(tt, "data.scaleway_block_volume.main"),
					checkBlockSnapshot(tt, "data.scaleway_block_volume.main", snapshotSpecsCheck{
						Size: scw.SizePtr(20 * scw.GB),
						Type: &sbsVolumeType,
					}),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_instance_server" "main" {
						name = "test-tf-action-instance-create-snapshot-sbs"
						type = "DEV1-S"
						image = "ubuntu_jammy"
						tags = [ "add", "tags", "to", "trigger", "update" ]
			
						root_volume {
							volume_type = "%s"
							size_in_gb = 20
						}
			
						lifecycle {
							action_trigger {
						  		events  = [after_update]
						  		actions = [action.scaleway_instance_create_snapshot.main]
							}
					  	}
					}

					data "scaleway_block_volume" "main" {
						volume_id = scaleway_instance_server.main.root_volume.0.volume_id
					}

					action "scaleway_instance_create_snapshot" "main" {
						config {
						  	volume_id = scaleway_instance_server.main.root_volume.0.volume_id
						  	tags = scaleway_instance_server.main.tags
						  	name = "custom-name-for-snapshot"
							wait = true
						}
					}`, sbsVolumeType),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					blocktestfuncs.IsVolumePresent(tt, "data.scaleway_block_volume.main"),
					checkBlockSnapshot(tt, "data.scaleway_block_volume.main", snapshotSpecsCheck{
						Name: scw.StringPtr("custom-name-for-snapshot"),
						Size: scw.SizePtr(20 * scw.GB),
						Tags: []string{"add", "tags", "to", "trigger", "update"},
					}),
				),
			},
			{
				// This step ensures that the snapshots are deleted before the volume they are based on,
				// otherwise they won't be returned by the ListSnapshot request.
				Config: fmt.Sprintf(`
					resource "scaleway_instance_server" "main" {
						name = "test-tf-action-instance-create-snapshot-sbs"
						type = "DEV1-S"
						image = "ubuntu_jammy"
						tags = [ "add", "tags", "to", "trigger", "update" ]
			
						root_volume {
							volume_type = "%s"
							size_in_gb = 20
						}
					}

					data "scaleway_block_volume" "main" {
						volume_id = scaleway_instance_server.main.root_volume.0.volume_id
					}`, sbsVolumeType),
				Check: destroyUntrackedBlockSnapshots(tt, "data.scaleway_block_volume.main"),
			},
		},
	})
}

func TestAccAction_InstanceCreateSnapshot_Scratch(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccAction_InstanceCreateSnapshot_Scratch because action are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	scratchVolumeType := instanceSDK.VolumeVolumeTypeScratch

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             instancechecks.IsServerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_instance_volume" "scratch" {
						name = "test-tf-action-instance-create-snapshot-scratch"
						type = "%s"
						size_in_gb = 50

					  	lifecycle {
							action_trigger {
						  		events  = [after_create]
						  		actions = [action.scaleway_instance_create_snapshot.scratch]
							}
					  	}
					}

					action "scaleway_instance_create_snapshot" "scratch" {
						config {
						  	volume_id = scaleway_instance_volume.scratch.id
							wait = true
						}
					}`, scratchVolumeType),
				ExpectError: regexp.MustCompile("Error when invoking action"), // scratch storage cannot be snapshot
			},
		},
	})
}

func TestAccAction_InstanceCreateSnapshot_Zone(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccAction_InstanceCreateSnapshot_Zone because action are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	localVolumeType := instanceSDK.VolumeVolumeTypeLSSD

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             instancechecks.IsServerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_instance_server" "main" {
						name = "test-tf-action-instance-create-snapshot-zone"
						type = "DEV1-S"
						image = "ubuntu_jammy"
						zone = "fr-par-2"

						root_volume {
							volume_type = "%s"
							size_in_gb = 20
						}

						lifecycle {
							action_trigger {
						  		events  = [after_create]
						  		actions = [action.scaleway_instance_create_snapshot.main]
							}
					  	}
					}

					data "scaleway_instance_volume" "main" {
						volume_id = scaleway_instance_server.main.root_volume.0.volume_id
						zone = "fr-par-2"
					}

					action "scaleway_instance_create_snapshot" "main" {
						config {
						  	volume_id = scaleway_instance_server.main.root_volume.0.volume_id
							wait = true
						}
					}`, localVolumeType),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsVolumePresent(tt, "data.scaleway_instance_volume.main"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "zone", "fr-par-2"),
					checkInstanceSnapshot(tt, "data.scaleway_instance_volume.main", snapshotSpecsCheck{
						Size: scw.SizePtr(20 * scw.GB),
						Type: &localVolumeType,
					}),
				),
			},
			{
				// This step ensures that the snapshots are deleted before the volume they are based on,
				// otherwise they won't be returned by the ListSnapshot request.
				Config: fmt.Sprintf(`
					resource "scaleway_instance_server" "main" {
						name = "test-tf-action-instance-create-snapshot-zone"
						type = "DEV1-S"
						image = "ubuntu_jammy"
						zone = "fr-par-2"

						root_volume {
							volume_type = "%s"
							size_in_gb = 20
						}
					}

					data "scaleway_instance_volume" "main" {
						volume_id = scaleway_instance_server.main.root_volume.0.volume_id
						zone = "fr-par-2"
					}`, localVolumeType),
				Check: destroyUntrackedInstanceSnapshots(tt, "data.scaleway_instance_volume.main"),
			},
		},
	})
}

func instanceSnapshotMatchesExpectedSpecs(snapshot instanceSDK.Snapshot, expected snapshotSpecsCheck) bool {
	if expected.Name != nil && *expected.Name != snapshot.Name {
		return false
	}

	if expected.Size != nil && *expected.Size != snapshot.Size {
		return false
	}

	if len(expected.Tags) > 0 && !reflect.DeepEqual(expected.Tags, snapshot.Tags) {
		return false
	}

	if len(snapshot.Tags) > len(expected.Tags) {
		return false
	}

	if expected.Type != nil && *expected.Type != snapshot.VolumeType {
		return false
	}

	return true
}

func checkInstanceSnapshot(tt *acctest.TestTools, n string, expectedSpecs snapshotSpecsCheck) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, zone, id, err := instance.NewAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		snapshots, err := api.ListSnapshots(&instanceSDK.ListSnapshotsRequest{
			Zone:         zone,
			BaseVolumeID: &id,
		}, scw.WithAllPages())
		if err != nil {
			return err
		}

		if snapshots.TotalCount == 0 {
			return fmt.Errorf("could not find any instance snapshot for volume %s", id)
		}

		for _, snapshot := range snapshots.Snapshots {
			if instanceSnapshotMatchesExpectedSpecs(*snapshot, expectedSpecs) {
				return nil
			}
		}

		return fmt.Errorf("could not find any instance snapshot that matches the specs %+v", expectedSpecs)
	}
}

func destroyUntrackedInstanceSnapshots(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, zone, id, err := instance.NewAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		snapshots, err := api.ListSnapshots(&instanceSDK.ListSnapshotsRequest{
			Zone:         zone,
			BaseVolumeID: &id,
		}, scw.WithAllPages())
		if err != nil {
			return err
		}

		for _, snapshot := range snapshots.Snapshots {
			err = api.DeleteSnapshot(&instanceSDK.DeleteSnapshotRequest{
				Zone:       zone,
				SnapshotID: snapshot.ID,
			})
			if err != nil {
				return err
			}
		}

		return nil
	}
}

func blockSnapshotMatchesExpectedSpecs(snapshot blockSDK.Snapshot, expected snapshotSpecsCheck) bool {
	if expected.Name != nil && *expected.Name != snapshot.Name {
		return false
	}

	if expected.Size != nil && *expected.Size != snapshot.Size {
		return false
	}

	if len(expected.Tags) > 0 && !reflect.DeepEqual(expected.Tags, snapshot.Tags) {
		return false
	}

	if len(snapshot.Tags) > len(expected.Tags) {
		return false
	}

	return true
}

func checkBlockSnapshot(tt *acctest.TestTools, n string, expectedSpecs snapshotSpecsCheck) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, zone, id, err := block.NewAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		snapshots, err := api.ListSnapshots(&blockSDK.ListSnapshotsRequest{
			Zone:     zone,
			VolumeID: &id,
		}, scw.WithAllPages())
		if err != nil {
			return err
		}

		if snapshots.TotalCount == 0 {
			return fmt.Errorf("could not find any block snapshot for volume %s", id)
		}

		for _, snapshot := range snapshots.Snapshots {
			if blockSnapshotMatchesExpectedSpecs(*snapshot, expectedSpecs) {
				return nil
			}
		}

		return fmt.Errorf("could not find any block snapshot that matches the specs %+v", expectedSpecs)
	}
}

func destroyUntrackedBlockSnapshots(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, zone, id, err := block.NewAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		snapshots, err := api.ListSnapshots(&blockSDK.ListSnapshotsRequest{
			Zone:     zone,
			VolumeID: &id,
		}, scw.WithAllPages())
		if err != nil {
			return err
		}

		for _, snapshot := range snapshots.Snapshots {
			err = api.DeleteSnapshot(&blockSDK.DeleteSnapshotRequest{
				Zone:       zone,
				SnapshotID: snapshot.ID,
			})
			if err != nil {
				return err
			}
		}

		return nil
	}
}
