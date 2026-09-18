package cockpit_test

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	cockpit "github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccActionCockpitGrafanaSyncDataSources_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccActionCockpitGrafanaSyncDataSources_Basic because actions are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_account_project" "project" {
						name = "tf_tests_cockpit_grafana_sync_data_sources"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					ensureCockpitGrafanaProvisioned(tt, "scaleway_account_project.project"),
				),
			},
			{
				Config: `
					resource "scaleway_account_project" "project" {
						name = "tf_tests_cockpit_grafana_sync_data_sources"
					}

					resource "scaleway_cockpit_source" "metrics" {
						project_id     = scaleway_account_project.project.id
						name           = "test-metrics-source"
						type           = "metrics"
						retention_days = 31

						lifecycle {
							action_trigger {
								events  = [after_create]
								actions = [action.scaleway_cockpit_grafana_sync_data_sources.main]
							}
						}
					}

					action "scaleway_cockpit_grafana_sync_data_sources" "main" {
						config {
							project_id = scaleway_account_project.project.id
						}
					}
				`,
			},
			{
				Config: `
					resource "scaleway_account_project" "project" {
						name = "tf_tests_cockpit_grafana_sync_data_sources"
					}

					resource "scaleway_cockpit_source" "metrics" {
						project_id     = scaleway_account_project.project.id
						name           = "test-metrics-source"
						type           = "metrics"
						retention_days = 31

						lifecycle {
							action_trigger {
								events  = [after_create]
								actions = [action.scaleway_cockpit_grafana_sync_data_sources.main]
							}
						}
					}

					action "scaleway_cockpit_grafana_sync_data_sources" "main" {
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

// ensureCockpitGrafanaProvisioned activates Grafana for a project by performing an
// IAM-authenticated request against the Grafana instance. Legacy CreateGrafanaUser
// no longer provisions Grafana; SyncGrafanaDataSources requires at least one access.
// GetGrafana goes through VCR; the Grafana /api/org call is live-only (outside api.scaleway.com).
func ensureCockpitGrafanaProvisioned(tt *acctest.TestTools, projectResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[projectResource]
		if !ok {
			return fmt.Errorf("not found: %s", projectResource)
		}

		projectID := rs.Primary.ID
		if projectID == "" {
			return fmt.Errorf("no ID set on %s", projectResource)
		}

		api := cockpit.NewGlobalAPI(tt.Meta.ScwClient())

		grafana, err := api.GetGrafana(&cockpit.GlobalAPIGetGrafanaRequest{
			ProjectID: projectID,
		}, scw.WithContext(tt.T.Context()))
		if err != nil {
			return fmt.Errorf("get grafana for project %s: %w", projectID, err)
		}

		if !*acctest.UpdateCassettes {
			return nil
		}

		secretKey, hasSecretKey := tt.Meta.ScwClient().GetSecretKey()
		if !hasSecretKey || secretKey == "" {
			return errors.New("missing secret key to provision grafana via IAM")
		}

		req, err := http.NewRequestWithContext(tt.T.Context(), http.MethodGet, grafana.GrafanaURL+"/api/org", nil)
		if err != nil {
			return err
		}

		req.Header.Set("X-Auth-Token", secretKey)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("access grafana for project %s: %w", projectID, err)
		}

		defer func() { _ = resp.Body.Close() }()

		_, _ = io.Copy(io.Discard, resp.Body)

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("access grafana for project %s: unexpected status %s", projectID, resp.Status)
		}

		return nil
	}
}
