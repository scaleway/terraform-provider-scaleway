package instance_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccActionServer_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
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
					}
				`,
				Check: resource.ComposeTestCheckFunc(),
			},
		},
	})
}

func TestAccActionServer_UnknownVerb(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
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
						  	action = "unknownVerb"
							server_id = scaleway_instance_server.main.id
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(),
			},
		},
	})
}
