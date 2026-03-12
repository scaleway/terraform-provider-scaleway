package instance_test

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	instanceSDK "github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/instance"
	instancechecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/instance/testfuncs"
)

type imageSpecsCheck struct {
	RootVolumeSize *scw.Size
	RootVolumeType *instanceSDK.VolumeVolumeType
}

func TestAccActionServer_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccActionServer_Basic because action are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             instancechecks.IsServerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_server" "main" {
						name = "test-terraform-action-server-basic"
						type = "DEV1-S"
						image = "ubuntu_jammy"

					  	lifecycle {
							action_trigger {
						  		events  = [after_create]
						  		actions = [action.scaleway_instance_server_action.main]
							}
					  	}
					}

					action "scaleway_instance_server_action" "main" {
						config {
						  	action = "reboot"
							server_id = scaleway_instance_server.main.id
						}
					}`,
			},
			{
				Config: `
					resource "scaleway_instance_server" "main" {
						name = "test-terraform-action-server-basic"
						type = "DEV1-S"
						image = "ubuntu_jammy"

					  	lifecycle {
							action_trigger {
						  		events  = [after_create]
						  		actions = [action.scaleway_instance_server_action.main]
							}
					  	}
					}

					action "scaleway_instance_server_action" "main" {
						config {
						  	action = "reboot"
							server_id = scaleway_instance_server.main.id
						}
					}

					data "scaleway_audit_trail_event" "instance" {
						resource_type = "instance_server"
						resource_id = scaleway_instance_server.main.id
						method_name = "ServerAction"
					}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scaleway_audit_trail_event.instance", "events.#"),
					func(state *terraform.State) error {
						rs, ok := state.RootModule().Resources["data.scaleway_audit_trail_event.instance"]
						if !ok {
							return errors.New("not found: data.scaleway_audit_trail_event.instance")
						}

						for key, value := range rs.Primary.Attributes {
							if !strings.Contains(key, "request_body") {
								continue
							}

							if value == `{"action":"reboot"}` {
								return nil
							}
						}

						return errors.New("did not found the reboot event")
					},
				),
			},
		},
	})
}

func TestAccActionServer_UnknownVerb(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccActionServer_UnknownVerb because action are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					action "scaleway_instance_server_action" "main" {
						config {
						  	action = "unknownVerb"
							server_id = "11111111-1111-1111-1111-111111111111"
						}
					}
				`,
				ExpectError: regexp.MustCompile("Invalid Attribute Value Match"),
			},
		},
	})
}

func TestAccActionServer_On_Off(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccActionServer_On_Off because action are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             instancechecks.IsServerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_instance_server" "main" {
						name = "test-terraform-action-server-on-off"
						type = "DEV1-S"
						image = "ubuntu_jammy"

					  	lifecycle {
							action_trigger {
						  		events  = [after_create]
						  		actions = [action.scaleway_instance_server_action.stop]
							}
					  	}
					}

					action "scaleway_instance_server_action" "stop" {
						config {
						  	action = "%s"
							server_id = scaleway_instance_server.main.id
							wait = true
						}
					}`, instanceSDK.ServerActionStopInPlace),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "state", instance.InstanceServerStateStandby),
					readActualServerState(tt, "scaleway_instance_server.main", instanceSDK.ServerStateStoppedInPlace.String()),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_instance_server" "main" {
						name = "should-be-powered-off"
						type = "DEV1-S"
						image = "ubuntu_jammy"

					  	lifecycle {
							action_trigger {
						  		events  = [after_update]
						  		actions = [action.scaleway_instance_server_action.poweroff]
							}
					  	}
					}

					action "scaleway_instance_server_action" "poweroff" {
						config {
						  	action = "%s"
							server_id = scaleway_instance_server.main.id
							wait = true
						}
					}`, instanceSDK.ServerActionPoweroff),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "state", instance.InstanceServerStateStopped),
					readActualServerState(tt, "scaleway_instance_server.main", instanceSDK.ServerStateStopped.String()),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_instance_server" "main" {
						name = "should-be-started"
						type = "DEV1-S"
						image = "ubuntu_jammy"

					  	lifecycle {
							action_trigger {
						  		events  = [after_update]
						  		actions = [action.scaleway_instance_server_action.start]
							}
					  	}
					}

					action "scaleway_instance_server_action" "start" {
						config {
						  	action = "%s"
							server_id = scaleway_instance_server.main.id
							wait = true
						}
					}`, instanceSDK.ServerActionPoweron),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "state", instance.InstanceServerStateStarted),
					readActualServerState(tt, "scaleway_instance_server.main", instanceSDK.ServerStateRunning.String()),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_instance_server" "main" {
						name = "should-be-terminated"
						type = "DEV1-S"
						image = "ubuntu_jammy"

					  	lifecycle {
							action_trigger {
						  		events  = [after_update]
						  		actions = [action.scaleway_instance_server_action.terminate]
							}
					  	}
					}

					action "scaleway_instance_server_action" "terminate" {
						config {
						  	action = "%s"
							server_id = scaleway_instance_server.main.id
							wait = false
						}
					}`, instanceSDK.ServerActionTerminate),
				ExpectNonEmptyPlan: true,
			},
			{
				RefreshState:       true,
				Check:              instancechecks.IsServerDestroyed(tt),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccActionServer_Backup(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccActionServer_Backup because action are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	rootVolumeType := instanceSDK.VolumeVolumeTypeLSSD
	rootVolumeSize := 15

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			instancechecks.IsServerDestroyed(tt),
			destroyUntrackedImagesAndSnapshots(tt, "scaleway_instance_server.main"),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_instance_server" "main" {
						name = "test-terraform-action-server-backup"
						type = "DEV1-S"
						image = "ubuntu_jammy"

						root_volume {
							volume_type = "%s"
							size_in_gb = %d
						}

					  	lifecycle {
							action_trigger {
						  		events  = [after_create]
						  		actions = [action.scaleway_instance_server_action.backup]
							}
					  	}
					}

					action "scaleway_instance_server_action" "backup" {
						config {
						  	action = "%s"
							server_id = scaleway_instance_server.main.id
							wait = true
						}
					}`, rootVolumeType, rootVolumeSize, instanceSDK.ServerActionBackup),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.main"),
					checkImage(tt, "scaleway_instance_server.main", imageSpecsCheck{
						RootVolumeSize: new(scw.Size(uint64(rootVolumeSize)) * scw.GB),
						RootVolumeType: &rootVolumeType,
					}),
				),
			},
		},
	})
}

func TestAccActionServer_Zone(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccActionServer_Zone because action are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             instancechecks.IsServerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_instance_server" "main" {
						name = "test-terraform-action-server-zone"
						type = "DEV1-S"
						image = "ubuntu_jammy"
						zone = "pl-waw-1"

					  	lifecycle {
							action_trigger {
						  		events  = [after_create]
						  		actions = [action.scaleway_instance_server_action.stop]
							}
					  	}
					}

					action "scaleway_instance_server_action" "stop" {
						config {
						  	action = "%s"
							server_id = scaleway_instance_server.main.id
							wait = true
						}
					}`, instanceSDK.ServerActionStopInPlace),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "zone", "pl-waw-1"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "state", instance.InstanceServerStateStandby),
					readActualServerState(tt, "scaleway_instance_server.main", instanceSDK.ServerStateStoppedInPlace.String()),
				),
			},
		},
	})
}

func readActualServerState(tt *acctest.TestTools, n string, expectedState string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, zone, id, err := instance.NewAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		server, err := api.GetServer(&instanceSDK.GetServerRequest{
			Zone:     zone,
			ServerID: id,
		})
		if err != nil {
			return err
		}

		if server.Server.State.String() != expectedState {
			return fmt.Errorf("expected server state to be %q, got %q", expectedState, server.Server.State)
		}

		return nil
	}
}

func checkImage(tt *acctest.TestTools, serverTFName string, expectedSpecs imageSpecsCheck) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[serverTFName]
		if !ok {
			return fmt.Errorf("resource not found: %s", serverTFName)
		}

		api, zone, _, err := instance.NewAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		serverName := rs.Primary.Attributes["name"]
		imageNamePrefix := "image_" + serverName

		images, err := api.ListImages(&instanceSDK.ListImagesRequest{
			Zone: zone,
			Name: &imageNamePrefix,
		}, scw.WithAllPages())
		if err != nil {
			return err
		}

		if images.TotalCount == 0 {
			return fmt.Errorf("could not find any image for server %s", serverName)
		} else if images.TotalCount > 1 {
			return fmt.Errorf("found multiple images for server %s", serverName)
		}

		image := images.Images[0]

		if expectedSpecs.RootVolumeSize != nil && *expectedSpecs.RootVolumeSize != image.RootVolume.Size {
			return fmt.Errorf("expected root volume size to be %d, got %d", expectedSpecs.RootVolumeSize, image.RootVolume.Size)
		}

		if expectedSpecs.RootVolumeType != nil && *expectedSpecs.RootVolumeType != image.RootVolume.VolumeType {
			return fmt.Errorf("expected root volume type to be %s, got %s", expectedSpecs.RootVolumeType, image.RootVolume.VolumeType)
		}

		return nil
	}
}

func destroyUntrackedImagesAndSnapshots(tt *acctest.TestTools, serverTFName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[serverTFName]
		if !ok {
			return fmt.Errorf("resource not found: %s", serverTFName)
		}

		api, zone, _, err := instance.NewAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		serverName := rs.Primary.Attributes["name"]
		imageNamePrefix := "image_" + serverName

		images, err := api.ListImages(&instanceSDK.ListImagesRequest{
			Zone: zone,
			Name: &imageNamePrefix,
		}, scw.WithAllPages())
		if err != nil {
			return err
		}

		for _, image := range images.Images {
			err = api.DeleteImage(&instanceSDK.DeleteImageRequest{
				Zone:    zone,
				ImageID: image.ID,
			})
			if err != nil {
				return err
			}

			err = api.DeleteSnapshot(&instanceSDK.DeleteSnapshotRequest{
				Zone:       zone,
				SnapshotID: image.RootVolume.ID,
			})
			if err != nil {
				return err
			}

			for _, extraVolume := range image.ExtraVolumes {
				err = api.DeleteSnapshot(&instanceSDK.DeleteSnapshotRequest{
					Zone:       zone,
					SnapshotID: extraVolume.ID,
				})
				if err != nil {
					return err
				}
			}
		}

		return nil
	}
}
