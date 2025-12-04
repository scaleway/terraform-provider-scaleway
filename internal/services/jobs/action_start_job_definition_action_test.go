package jobs_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccActionJobDefinitionStart_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccActionJobDefinitionStart_Basic because actions are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_job_definition" "main" {
						name         = "test-jobs-action-start"
						cpu_limit    = 120
						memory_limit = 256
						image_uri    = "docker.io/alpine:latest"
						command      = "echo 'Hello World'"

						lifecycle {
							action_trigger {
								events  = [after_create]
								actions = [action.scaleway_job_definition_start_action.main]
							}
						}
					}

					action "scaleway_job_definition_start_action" "main" {
						config {
							job_definition_id = scaleway_job_definition.main.id
						}
					}
				`,
			},
			{
				Config: `
					resource "scaleway_job_definition" "main" {
						name         = "test-jobs-action-start"
						cpu_limit    = 120
						memory_limit = 256
						image_uri    = "docker.io/alpine:latest"
						command      = "echo 'Hello World'"

						lifecycle {
							action_trigger {
								events  = [after_create]
								actions = [action.scaleway_job_definition_start_action.main]
							}
						}
					}

					action "scaleway_job_definition_start_action" "main" {
						config {
							job_definition_id = scaleway_job_definition.main.id
						}
					}

					data "scaleway_audit_trail_event" "jobs" {
						resource_type = "job_definition"
						resource_id   = scaleway_job_definition.main.id
						method_name    = "StartJobDefinition"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scaleway_audit_trail_event.jobs", "events.#"),
					func(state *terraform.State) error {
						rs, ok := state.RootModule().Resources["data.scaleway_audit_trail_event.jobs"]
						if !ok {
							return errors.New("not found: data.scaleway_audit_trail_event.jobs")
						}

						for key, value := range rs.Primary.Attributes {
							if !strings.Contains(key, "method_name") {
								continue
							}

							if value == "StartJobDefinition" {
								return nil
							}
						}

						return errors.New("did not find the StartJobDefinition event")
					},
				),
			},
		},
	})
}

