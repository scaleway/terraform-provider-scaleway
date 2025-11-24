package instance_test

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccActionServer_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_server" "main" {
						name = "test-terraform-datasource-private-nic"
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
						name = "test-terraform-datasource-private-nic"
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

						countStr := rs.Primary.Attributes["events.#"]

						count, err := strconv.Atoi(countStr)
						if err != nil {
							return fmt.Errorf("could not parse events.# as integer: %w", err)
						}

						if count < 1 {
							return fmt.Errorf("expected events count > 1, got %d", count)
						}

						return nil
					},
				),
			},
		},
	})
}

func TestAccActionServer_UnknownVerb(t *testing.T) {
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
