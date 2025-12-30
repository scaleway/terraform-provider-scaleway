package baremetal_test

import (
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	baremetalchecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/baremetal/testfuncs"
)

func TestAccBaremetalServerAction_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccBaremetalServerAction_Basic because actions are not yet supported on OpenTofu")
	}

	if !IsOfferAvailable(OfferName, scw.Zone(Zone), tt) {
		t.Skip("Skipping TestAccBaremetalServerAction_Basic because offer is out of stock")
	}

	sshKeyName := "TestAccBaremetalServerAction_Basic"
	serverName := "TestAccBaremetalServerAction_Basic"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             baremetalchecks.CheckServerDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				data "scaleway_baremetal_os" "my_os" {
				  zone    = "%s"
				  name    = "Ubuntu"
				  version = "22.04 LTS (Jammy Jellyfish)"
				}
				
				resource "scaleway_iam_ssh_key" "main" {
				  name       = "%s"
				  public_key = "%s"
				}
				
				resource "scaleway_baremetal_server" "base" {
				  name        = "%s"
				  zone        = "%s"
				  description = "initial-description"
				  offer       = "%s"
				  os          = data.scaleway_baremetal_os.my_os.os_id
				
				  tags = [
					"terraform-test",
					"scaleway_baremetal_server",
					"minimal",
				  ]
	
					lifecycle {
					action_trigger {
					  events = [after_update]
					  actions = [
						action.scaleway_baremetal_server_action.base_stop,
						action.scaleway_baremetal_server_action.base_start,
						action.scaleway_baremetal_server_action.base_reboot,
					  ]
					}
				  }
					
				  ssh_key_ids = [scaleway_iam_ssh_key.main.id]
				}
				
				action "scaleway_baremetal_server_action" "base_stop" {
				  config {
					action    = "stop"
					server_id = scaleway_baremetal_server.base.id
					wait      = true
				  }
				}
				
				action "scaleway_baremetal_server_action" "base_start" {
				  config {
					action    = "start"
					server_id = scaleway_baremetal_server.base.id
					wait      = true
				  }
				}
				
				action "scaleway_baremetal_server_action" "base_reboot" {
				  config {
					action    = "reboot"
					server_id = scaleway_baremetal_server.base.id
					boot_type = "normal"
					wait      = true
				  }
				}
					`, Zone, sshKeyName, SSHKeyBaremetal, serverName, Zone, OfferName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBaremetalServerExists(tt, "scaleway_baremetal_server.base"),
				),
			},
			{
				Config: fmt.Sprintf(`
				data "scaleway_baremetal_os" "my_os" {
				  zone    = "%s"
				  name    = "Ubuntu"
				  version = "22.04 LTS (Jammy Jellyfish)"
				}
				
				resource "scaleway_iam_ssh_key" "main" {
				  name       = "%s"
				  public_key = "%s"
				}
				
				resource "scaleway_baremetal_server" "base" {
				  name        = "%s"
				  zone        = "%s"
				  description = "updated-description"
				  offer       = "%s"
				  os          = data.scaleway_baremetal_os.my_os.os_id
				
				  tags = [
					"terraform-test",
					"scaleway_baremetal_server",
					"minimal",
				  ]
				
				  ssh_key_ids = [scaleway_iam_ssh_key.main.id]
				
				  lifecycle {
					action_trigger {
					  events = [after_update]
					  actions = [
						action.scaleway_baremetal_server_action.base_stop,
						action.scaleway_baremetal_server_action.base_start,
						action.scaleway_baremetal_server_action.base_reboot,
					  ]
					}
				  }
				}
				
				action "scaleway_baremetal_server_action" "base_stop" {
				  config {
					action    = "stop"
					server_id = scaleway_baremetal_server.base.id
					wait      = true
				  }
				}
				
				action "scaleway_baremetal_server_action" "base_start" {
				  config {
					action    = "start"
					server_id = scaleway_baremetal_server.base.id
					wait      = true
				  }
				}
				
				action "scaleway_baremetal_server_action" "base_reboot" {
				  config {
					action    = "reboot"
					server_id = scaleway_baremetal_server.base.id
					boot_type = "normal"
					wait      = true
				  }
				}
					`, Zone, sshKeyName, SSHKeyBaremetal, serverName, Zone, OfferName),
			},
			{
				Config: fmt.Sprintf(`
					data "scaleway_baremetal_os" "my_os" {
					  zone    = "%s"
					  name    = "Ubuntu"
					  version = "22.04 LTS (Jammy Jellyfish)"
					}
					
					resource "scaleway_iam_ssh_key" "main" {
					  name       = "%s"
					  public_key = "%s"
					}
					
					resource "scaleway_baremetal_server" "base" {
					  name        = "%s"
					  zone        = "%s"
					  description = "updated-description"
					  offer       = "%s"
					  os          = data.scaleway_baremetal_os.my_os.os_id
					
					  tags = [
						"terraform-test",
						"scaleway_baremetal_server",
						"minimal",
					  ]
					
					  ssh_key_ids = [scaleway_iam_ssh_key.main.id]
					
					  lifecycle {
						action_trigger {
						  events = [after_update]
						  actions = [
							action.scaleway_baremetal_server_action.base_stop,
							action.scaleway_baremetal_server_action.base_start,
							action.scaleway_baremetal_server_action.base_reboot,
						  ]
						}
					  }
					}
					
					action "scaleway_baremetal_server_action" "base_stop" {
					  config {
						action    = "stop"
						server_id = scaleway_baremetal_server.base.id
						wait      = true
					  }
					}
					
					action "scaleway_baremetal_server_action" "base_start" {
					  config {
						action    = "start"
						server_id = scaleway_baremetal_server.base.id
						wait      = true
					  }
					}
					
					action "scaleway_baremetal_server_action" "base_reboot" {
					  config {
						action    = "reboot"
						server_id = scaleway_baremetal_server.base.id
						boot_type = "normal"
						wait      = true
					  }
					}
					
					data "scaleway_audit_trail_event" "stop" {
					  resource_id = scaleway_baremetal_server.base.id
					  method_name = "StopServer"
					}
					
					data "scaleway_audit_trail_event" "start" {
					  resource_id = scaleway_baremetal_server.base.id
					  method_name = "StartServer"
					}
					
					data "scaleway_audit_trail_event" "reboot" {
					  resource_id = scaleway_baremetal_server.base.id
					  method_name = "RebootServer"
					}
					`, Zone, sshKeyName, SSHKeyBaremetal, serverName, Zone, OfferName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBaremetalServerExists(tt, "scaleway_baremetal_server.base"),

					resource.TestCheckResourceAttrSet("data.scaleway_audit_trail_event.stop", "events.#"),
					resource.TestCheckResourceAttrSet("data.scaleway_audit_trail_event.start", "events.#"),
					resource.TestCheckResourceAttrSet("data.scaleway_audit_trail_event.reboot", "events.#"),

					checkEventsOccurrence("data.scaleway_audit_trail_event.stop"),
					checkEventsOccurrence("data.scaleway_audit_trail_event.start"),
					checkEventsOccurrence("data.scaleway_audit_trail_event.reboot"),
				),
			},
		},
	})
}

func checkEventsOccurrence(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return errors.New("not found: " + resourceName)
		}

		countStr := rs.Primary.Attributes["events.#"]

		count, err := strconv.Atoi(countStr)
		if err != nil {
			return fmt.Errorf("could not parse events.# as integer: %w", err)
		}

		if count != 1 {
			return fmt.Errorf("expected exactly 1 event, got %d", count)
		}

		return nil
	}
}
