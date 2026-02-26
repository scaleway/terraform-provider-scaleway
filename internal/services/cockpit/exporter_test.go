package cockpit_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	cockpitSDK "github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/cockpit"
)

const defaultOrgIDPlaceholder = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"

func TestAccCockpitExporter_Basic_Datadog(t *testing.T) {
	if *acctest.UpdateCassettes {
		t.Cleanup(func() { _ = acctest.AnonymizeCassetteForTest(t, "") })
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	orgID, orgIDExists := tt.Meta.ScwClient().GetDefaultOrganizationID()
	if !orgIDExists {
		orgID = defaultOrgIDPlaceholder
	}

	exporterName := "my-datadog-exporter"
	datadogAPIKey := "00000000000000000000000000000000"

	if k := os.Getenv("TF_TEST_DATADOG_API_KEY"); k != "" {
		datadogAPIKey = k
	}

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isExporterDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					data "scaleway_account_project" "project" {
						name             = "default"
						organization_id = "%s"
					}

					data "scaleway_cockpit_sources" "scaleway_metrics" {
						project_id = data.scaleway_account_project.project.id
						origin     = "scaleway"
						type       = "metrics"
					}

					resource "scaleway_cockpit_exporter" "main" {
						project_id        = data.scaleway_account_project.project.id
						datasource_id     = data.scaleway_cockpit_sources.scaleway_metrics.sources[0].id
						name              = "%s"
						exported_products = ["all"]
						datadog_destination {
							api_key  = "%s"
							endpoint = "https://api.datadoghq.com"
						}
					}
				`, orgID, exporterName, datadogAPIKey),
				Check: resource.ComposeTestCheckFunc(
					isExporterPresent(tt, "scaleway_cockpit_exporter.main"),
					resource.TestCheckResourceAttr("scaleway_cockpit_exporter.main", "name", exporterName),
					resource.TestCheckResourceAttr("scaleway_cockpit_exporter.main", "region", "fr-par"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit_exporter.main", "status"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit_exporter.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit_exporter.main", "updated_at"),
					resource.TestCheckResourceAttrPair("scaleway_cockpit_exporter.main", "project_id", "data.scaleway_account_project.project", "id"),
				),
			},
		},
	})
}

func TestAccCockpitExporter_Basic_OTLP(t *testing.T) {
	if *acctest.UpdateCassettes {
		t.Cleanup(func() { _ = acctest.AnonymizeCassetteForTest(t, "") })
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	orgID, orgIDExists := tt.Meta.ScwClient().GetDefaultOrganizationID()
	if !orgIDExists {
		orgID = defaultOrgIDPlaceholder
	}

	exporterName := "my-exporter"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isExporterDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					data "scaleway_account_project" "project" {
						name             = "default"
						organization_id = "%s"
					}

					data "scaleway_cockpit_sources" "scaleway_metrics" {
						project_id = data.scaleway_account_project.project.id
						origin     = "scaleway"
						type       = "metrics"
					}

					resource "scaleway_cockpit_source" "otlp_target" {
						project_id     = data.scaleway_account_project.project.id
						name           = "otlp-exporter-target"
						type           = "metrics"
						retention_days = 31
					}

					resource "scaleway_cockpit_exporter" "main" {
						project_id        = data.scaleway_account_project.project.id
						datasource_id     = data.scaleway_cockpit_sources.scaleway_metrics.sources[0].id
						name              = "%s"
						exported_products = ["all"]
						otlp_destination {
							endpoint = scaleway_cockpit_source.otlp_target.push_url
						}
					}
				`, orgID, exporterName),
				Check: resource.ComposeTestCheckFunc(
					isExporterPresent(tt, "scaleway_cockpit_exporter.main"),
					resource.TestCheckResourceAttr("scaleway_cockpit_exporter.main", "name", exporterName),
					resource.TestCheckResourceAttr("scaleway_cockpit_exporter.main", "region", "fr-par"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit_exporter.main", "status"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit_exporter.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit_exporter.main", "updated_at"),
					resource.TestCheckResourceAttrPair("scaleway_cockpit_exporter.main", "project_id", "data.scaleway_account_project.project", "id"),
				),
			},
		},
	})
}

func TestAccCockpitExporter_Update(t *testing.T) {
	if *acctest.UpdateCassettes {
		t.Cleanup(func() { _ = acctest.AnonymizeCassetteForTest(t, "") })
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	orgID, orgIDExists := tt.Meta.ScwClient().GetDefaultOrganizationID()
	if !orgIDExists {
		orgID = defaultOrgIDPlaceholder
	}

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isExporterDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					data "scaleway_account_project" "project" {
						name             = "default"
						organization_id = "%s"
					}

					data "scaleway_cockpit_sources" "scaleway_metrics" {
						project_id = data.scaleway_account_project.project.id
						origin     = "scaleway"
						type       = "metrics"
					}

					resource "scaleway_cockpit_source" "otlp_target" {
						project_id     = data.scaleway_account_project.project.id
						name           = "otlp-exporter-update-target"
						type           = "metrics"
						retention_days = 31
					}

					resource "scaleway_cockpit_exporter" "main" {
						project_id        = data.scaleway_account_project.project.id
						datasource_id     = data.scaleway_cockpit_sources.scaleway_metrics.sources[0].id
						name              = "initial-name"
						exported_products = ["all"]
						otlp_destination {
							endpoint = scaleway_cockpit_source.otlp_target.push_url
						}
					}
				`, orgID),
				Check: resource.ComposeTestCheckFunc(
					isExporterPresent(tt, "scaleway_cockpit_exporter.main"),
					resource.TestCheckResourceAttr("scaleway_cockpit_exporter.main", "name", "initial-name"),
				),
			},
			{
				Config: fmt.Sprintf(`
					data "scaleway_account_project" "project" {
						name             = "default"
						organization_id = "%s"
					}

					data "scaleway_cockpit_sources" "scaleway_metrics" {
						project_id = data.scaleway_account_project.project.id
						origin     = "scaleway"
						type       = "metrics"
					}

					resource "scaleway_cockpit_source" "otlp_target" {
						project_id     = data.scaleway_account_project.project.id
						name           = "otlp-exporter-update-target"
						type           = "metrics"
						retention_days = 31
					}

					resource "scaleway_cockpit_exporter" "main" {
						project_id        = data.scaleway_account_project.project.id
						datasource_id     = data.scaleway_cockpit_sources.scaleway_metrics.sources[0].id
						name              = "updated-name"
						description      = "Updated description"
						exported_products = ["all"]
						otlp_destination {
							endpoint = scaleway_cockpit_source.otlp_target.push_url
						}
					}
				`, orgID),
				Check: resource.ComposeTestCheckFunc(
					isExporterPresent(tt, "scaleway_cockpit_exporter.main"),
					resource.TestCheckResourceAttr("scaleway_cockpit_exporter.main", "name", "updated-name"),
					resource.TestCheckResourceAttr("scaleway_cockpit_exporter.main", "description", "Updated description"),
					resource.TestCheckResourceAttr("scaleway_cockpit_exporter.main", "exported_products.0", "all"),
				),
			},
			{
				Config: fmt.Sprintf(`
					data "scaleway_account_project" "project" {
						name             = "default"
						organization_id = "%s"
					}

					data "scaleway_cockpit_sources" "scaleway_metrics" {
						project_id = data.scaleway_account_project.project.id
						origin     = "scaleway"
						type       = "metrics"
					}

					resource "scaleway_cockpit_source" "otlp_target" {
						project_id     = data.scaleway_account_project.project.id
						name           = "otlp-exporter-update-target"
						type           = "metrics"
						retention_days = 31
					}

					resource "scaleway_cockpit_exporter" "main" {
						project_id        = data.scaleway_account_project.project.id
						datasource_id     = data.scaleway_cockpit_sources.scaleway_metrics.sources[0].id
						name              = "updated-name"
						description      = "Updated description"
						exported_products = ["lb", "object-storage", "rdb"]
						otlp_destination {
							endpoint = scaleway_cockpit_source.otlp_target.push_url
						}
					}
				`, orgID),
				Check: resource.ComposeTestCheckFunc(
					isExporterPresent(tt, "scaleway_cockpit_exporter.main"),
					resource.TestCheckResourceAttr("scaleway_cockpit_exporter.main", "name", "updated-name"),
					resource.TestCheckResourceAttr("scaleway_cockpit_exporter.main", "exported_products.0", "lb"),
					resource.TestCheckResourceAttr("scaleway_cockpit_exporter.main", "exported_products.1", "object-storage"),
					resource.TestCheckResourceAttr("scaleway_cockpit_exporter.main", "exported_products.2", "rdb"),
				),
			},
		},
	})
}

func isExporterPresent(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource cockpit exporter not found: %s", n)
		}

		api, region, id, err := cockpit.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetExporter(&cockpitSDK.RegionalAPIGetExporterRequest{
			Region:     region,
			ExporterID: id,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func isExporterDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		ctx := context.Background()

		return retry.RetryContext(ctx, DestroyWaitTimeout, func() *retry.RetryError {
			for _, rs := range state.RootModule().Resources {
				if rs.Type != "scaleway_cockpit_exporter" {
					continue
				}

				api, region, id, err := cockpit.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
				if err != nil {
					return retry.NonRetryableError(err)
				}

				_, err = api.GetExporter(&cockpitSDK.RegionalAPIGetExporterRequest{
					Region:     region,
					ExporterID: id,
				})

				switch {
				case err == nil:
					return retry.RetryableError(fmt.Errorf("cockpit exporter (%s) still exists", rs.Primary.ID))
				case httperrors.Is404(err):
					continue
				default:
					return retry.NonRetryableError(err)
				}
			}

			return nil
		})
	}
}
