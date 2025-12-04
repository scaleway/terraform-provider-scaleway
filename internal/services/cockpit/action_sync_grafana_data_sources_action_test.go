package cockpit_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccActionCockpitSyncGrafanaDataSources_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccActionCockpitSyncGrafanaDataSources_Basic because actions are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_account_project" "project" {
						name = "tf_tests_cockpit_sync_data_sources"
					}

					resource "scaleway_cockpit_grafana_user" "main" {
						project_id = scaleway_account_project.project.id
						login      = "test-sync-user"
						role       = "viewer"
					}

					resource "scaleway_cockpit_source" "metrics" {
						project_id     = scaleway_account_project.project.id
						name           = "test-metrics-source"
						type           = "metrics"
						retention_days = 31

						lifecycle {
							action_trigger {
								events  = [after_create]
								actions = [action.scaleway_cockpit_sync_grafana_data_sources_action.main]
							}
						}

						depends_on = [scaleway_cockpit_grafana_user.main]
					}

					action "scaleway_cockpit_sync_grafana_data_sources_action" "main" {
						config {
							project_id = scaleway_account_project.project.id
						}
					}
				`,
			},
			{
				Config: `
					resource "scaleway_account_project" "project" {
						name = "tf_tests_cockpit_sync_data_sources"
					}

					resource "scaleway_cockpit_grafana_user" "main" {
						project_id = scaleway_account_project.project.id
						login      = "test-sync-user"
						role       = "viewer"
					}

					resource "scaleway_cockpit_source" "metrics" {
						project_id     = scaleway_account_project.project.id
						name           = "test-metrics-source"
						type           = "metrics"
						retention_days = 31

						lifecycle {
							action_trigger {
								events  = [after_create]
								actions = [action.scaleway_cockpit_sync_grafana_data_sources_action.main]
							}
						}

						depends_on = [scaleway_cockpit_grafana_user.main]
					}

					action "scaleway_cockpit_sync_grafana_data_sources_action" "main" {
						config {
							project_id = scaleway_account_project.project.id
						}
					}

					data "scaleway_audit_trail_event" "cockpit" {
						project_id  = scaleway_account_project.project.id
						method_name = "SyncGrafanaDataSources"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scaleway_audit_trail_event.cockpit", "events.#"),
					func(state *terraform.State) error {
						rs, ok := state.RootModule().Resources["data.scaleway_audit_trail_event.cockpit"]
						if !ok {
							return errors.New("not found: data.scaleway_audit_trail_event.cockpit")
						}

						for key, value := range rs.Primary.Attributes {
							if !strings.Contains(key, "method_name") {
								continue
							}

							if value == "SyncGrafanaDataSources" {
								return nil
							}
						}

						return errors.New("did not find the SyncGrafanaDataSources event")
					},
				),
			},
		},
	})
}

