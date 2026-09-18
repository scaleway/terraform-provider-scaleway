package cockpit_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccActionCockpitActivateGrafana_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccActionCockpitActivateGrafana_Basic because actions are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_account_project" "project" {
						name = "tf_tests_cockpit_activate_grafana"

						lifecycle {
							action_trigger {
								events  = [after_create]
								actions = [action.scaleway_cockpit_activate_grafana.main]
							}
						}
					}

					action "scaleway_cockpit_activate_grafana" "main" {
						config {
							project_id = scaleway_account_project.project.id
						}
					}
				`,
			},
		},
	})
}
