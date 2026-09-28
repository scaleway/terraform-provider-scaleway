package cockpit_test

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	cockpitSDK "github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/cockpit"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
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

					resource "scaleway_cockpit" "main" {
						project_id = scaleway_account_project.project.id
					}

					data "scaleway_cockpit_grafana" "main" {
						project_id = scaleway_cockpit.main.project_id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_grafana.main", "grafana_url"),
					activateGrafanaViaIAMCheck(tt, "scaleway_account_project.project"),
				),
			},
			{
				Config: `
					resource "scaleway_account_project" "project" {
						name = "tf_tests_cockpit_grafana_sync_data_sources"
					}

					resource "scaleway_cockpit" "main" {
						project_id = scaleway_account_project.project.id
					}

					data "scaleway_cockpit_grafana" "main" {
						project_id = scaleway_cockpit.main.project_id
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

						depends_on = [data.scaleway_cockpit_grafana.main]
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

					resource "scaleway_cockpit" "main" {
						project_id = scaleway_account_project.project.id
					}

					data "scaleway_cockpit_grafana" "main" {
						project_id = scaleway_cockpit.main.project_id
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

						depends_on = [data.scaleway_cockpit_grafana.main]
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

// activateGrafanaViaIAMCheck provisions Grafana via IAM first-access (replaces CreateGrafanaUser).
func activateGrafanaViaIAMCheck(tt *acctest.TestTools, projectResource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[projectResource]
		if !ok {
			return fmt.Errorf("not found: %s", projectResource)
		}

		projectID := rs.Primary.ID
		if projectID == "" {
			return fmt.Errorf("empty project id for %s", projectResource)
		}

		api, err := cockpit.NewGlobalAPI(tt.Meta)
		if err != nil {
			return err
		}

		ctx := tt.T.Context()

		grafana, err := api.GetGrafana(&cockpitSDK.GlobalAPIGetGrafanaRequest{
			ProjectID: projectID,
		}, scw.WithContext(ctx))
		if err != nil {
			return fmt.Errorf("get grafana: %w", err)
		}

		if grafana == nil || grafana.GrafanaURL == "" {
			return fmt.Errorf("empty grafana URL for project %s", projectID)
		}

		secretKey, hasSecretKey := tt.Meta.ScwClient().GetSecretKey()
		if !hasSecretKey || secretKey == "" {
			return errors.New("missing secret key to activate grafana via IAM")
		}

		const (
			maxAttempts      = 5
			defaultRetryWait = 5 * time.Second
		)

		retryWait := defaultRetryWait
		if transport.DefaultWaitRetryInterval != nil {
			retryWait = *transport.DefaultWaitRetryInterval
		}

		var lastStatus string

		for attempt := 1; attempt <= maxAttempts; attempt++ {
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, grafana.GrafanaURL+"/api/org", nil)
			if err != nil {
				return err
			}

			req.Header.Set("X-Auth-Token", secretKey)

			httpResp, err := tt.Meta.HTTPClient().Do(req)
			if err != nil {
				return fmt.Errorf("access grafana: %w", err)
			}

			_, _ = io.Copy(io.Discard, httpResp.Body)
			_ = httpResp.Body.Close()

			if httpResp.StatusCode == http.StatusOK {
				return nil
			}

			lastStatus = httpResp.Status

			if httpResp.StatusCode < http.StatusInternalServerError || attempt == maxAttempts {
				break
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(retryWait):
			}
		}

		return fmt.Errorf("access grafana: unexpected status %s", lastStatus)
	}
}
