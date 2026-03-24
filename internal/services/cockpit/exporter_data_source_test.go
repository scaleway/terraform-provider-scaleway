package cockpit_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccCockpitExporter_DataSource_ByID(t *testing.T) {
	if *acctest.UpdateCassettes {
		t.Cleanup(func() { _ = acctest.AnonymizeCassetteForTest(t, "") })
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	orgID, orgIDExists := tt.Meta.ScwClient().GetDefaultOrganizationID()
	if !orgIDExists {
		orgID = defaultOrgIDPlaceholder
	}

	exporterName := "ds-exporter-by-id"

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
						name           = "otlp-ds-by-id-target"
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

					data "scaleway_cockpit_exporter" "by_id" {
						id = scaleway_cockpit_exporter.main.id
					}
				`, orgID, exporterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.scaleway_cockpit_exporter.by_id", "id", "scaleway_cockpit_exporter.main", "id"),
					resource.TestCheckResourceAttr("data.scaleway_cockpit_exporter.by_id", "name", exporterName),
					resource.TestCheckResourceAttr("data.scaleway_cockpit_exporter.by_id", "region", "fr-par"),
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_exporter.by_id", "status"),
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_exporter.by_id", "project_id"),
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_exporter.by_id", "datasource_id"),
					resource.TestCheckResourceAttrPair("data.scaleway_cockpit_exporter.by_id", "project_id", "data.scaleway_account_project.project", "id"),
				),
			},
		},
	})
}

func TestAccCockpitExporter_DataSource_ByName(t *testing.T) {
	if *acctest.UpdateCassettes {
		t.Cleanup(func() { _ = acctest.AnonymizeCassetteForTest(t, "") })
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	orgID, orgIDExists := tt.Meta.ScwClient().GetDefaultOrganizationID()
	if !orgIDExists {
		orgID = defaultOrgIDPlaceholder
	}

	exporterName := "ds-exporter-by-name"

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
						name           = "otlp-ds-by-name-target"
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

					data "scaleway_cockpit_exporter" "by_name" {
						project_id = data.scaleway_account_project.project.id
						name       = "%s"
						depends_on = [scaleway_cockpit_exporter.main]
					}
				`, orgID, exporterName, exporterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.scaleway_cockpit_exporter.by_name", "id", "scaleway_cockpit_exporter.main", "id"),
					resource.TestCheckResourceAttr("data.scaleway_cockpit_exporter.by_name", "name", exporterName),
					resource.TestCheckResourceAttr("data.scaleway_cockpit_exporter.by_name", "region", "fr-par"),
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_exporter.by_name", "status"),
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_exporter.by_name", "project_id"),
					resource.TestCheckResourceAttrPair("data.scaleway_cockpit_exporter.by_name", "project_id", "data.scaleway_account_project.project", "id"),
				),
			},
		},
	})
}
