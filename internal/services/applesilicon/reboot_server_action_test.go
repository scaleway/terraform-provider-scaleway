package applesilicon_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/audittrail"
)

func TestAccRebootServer_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isServerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_apple_silicon_server main {
						name = "TestAccServerBasic"
						type = "M4-M"
						public_bandwidth = 1000000000

						lifecycle {
							action_trigger {
							  events = [after_create]
							  actions = [
								action.scaleway_apple_silicon_reboot_server_action.main_reboot,
							  ]
							}
						}
					}

					action "scaleway_apple_silicon_reboot_server_action" "main_reboot" {
					  config {
						server_id = scaleway_apple_silicon_server.main.id
						wait      = true
					  }
					}
				`,
			},
			{
				Config: `
					resource scaleway_apple_silicon_server main {
						name = "TestAccServerBasic"
						type = "M4-M"
						public_bandwidth = 1000000000

						lifecycle {
							action_trigger {
							  events = [after_create]
							  actions = [
								action.scaleway_apple_silicon_reboot_server_action.main_reboot,
							  ]
							}
						}
					}

					action scaleway_apple_silicon_reboot_server_action main_reboot {
					  config {
						server_id = scaleway_apple_silicon_server.main.id
						wait      = true
					  }
					}

					data scaleway_audit_trail_event reboot {
						resource_id = scaleway_apple_silicon_server.main.id
					  	method_name = "RebootServer"
					}

				`,
				Check: resource.ComposeTestCheckFunc(
					isServerPresent(tt, "scaleway_apple_silicon_server.main"),
					resource.TestCheckResourceAttrSet("data.scaleway_audit_trail_event.reboot", "events.#"),
					audittrail.CheckEventsOccurrence("data.scaleway_audit_trail_event.reboot"),
				),
			},
		},
	})
}
