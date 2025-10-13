package cockpit_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccCockpitSources_DataSource_All(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isSourceDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_account_project" "project" {
						name = "tf_tests_cockpit_sources_all"
				  	}

					resource "scaleway_cockpit_source" "metrics" {
					  project_id = scaleway_account_project.project.id
					  name       = "test-metrics-source"
					  type       = "metrics"
					  retention_days = 30
					}

					resource "scaleway_cockpit_source" "logs" {
					  project_id = scaleway_account_project.project.id
					  name       = "test-logs-source"
					  type       = "logs"
					  retention_days = 30
					}

					data "scaleway_cockpit_sources" "all" {
					  project_id = scaleway_account_project.project.id
					  depends_on = [scaleway_cockpit_source.metrics, scaleway_cockpit_source.logs]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaleway_cockpit_sources.all", "sources.#", "2"),
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_sources.all", "sources.0.id"),
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_sources.all", "sources.1.id"),
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_sources.all", "sources.0.url"),
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_sources.all", "sources.1.url"),
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_sources.all", "sources.0.push_url"),
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_sources.all", "sources.1.push_url"),
				),
			},
		},
	})
}

func TestAccCockpitSources_DataSource_ByType(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isSourceDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_account_project" "project" {
						name = "tf_tests_cockpit_sources_by_type"
				  	}

					resource "scaleway_cockpit_source" "metrics" {
					  project_id = scaleway_account_project.project.id
					  name       = "test-metrics-source"
					  type       = "metrics"
					  retention_days = 30
					}

					resource "scaleway_cockpit_source" "logs" {
					  project_id = scaleway_account_project.project.id
					  name       = "test-logs-source"
					  type       = "logs"
					  retention_days = 30
					}

					data "scaleway_cockpit_sources" "metrics_only" {
					  project_id = scaleway_account_project.project.id
					  type       = "metrics"
					  depends_on = [scaleway_cockpit_source.metrics, scaleway_cockpit_source.logs]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaleway_cockpit_sources.metrics_only", "sources.#", "1"),
					resource.TestCheckResourceAttr("data.scaleway_cockpit_sources.metrics_only", "sources.0.type", "metrics"),
					resource.TestCheckResourceAttr("data.scaleway_cockpit_sources.metrics_only", "sources.0.name", "test-metrics-source"),
				),
			},
		},
	})
}

func TestAccCockpitSources_DataSource_ByName(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isSourceDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_account_project" "project" {
						name = "tf_tests_cockpit_sources_by_name"
				  	}

					resource "scaleway_cockpit_source" "metrics" {
					  project_id = scaleway_account_project.project.id
					  name       = "test-metrics-source"
					  type       = "metrics"
					  retention_days = 30
					}

					resource "scaleway_cockpit_source" "logs" {
					  project_id = scaleway_account_project.project.id
					  name       = "test-logs-source"
					  type       = "logs"
					  retention_days = 30
					}

					data "scaleway_cockpit_sources" "by_name" {
					  project_id = scaleway_account_project.project.id
					  name       = "test-metrics-source"
					  depends_on = [scaleway_cockpit_source.metrics, scaleway_cockpit_source.logs]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaleway_cockpit_sources.by_name", "sources.#", "1"),
					resource.TestCheckResourceAttr("data.scaleway_cockpit_sources.by_name", "sources.0.name", "test-metrics-source"),
					resource.TestCheckResourceAttr("data.scaleway_cockpit_sources.by_name", "sources.0.type", "metrics"),
				),
			},
		},
	})
}

func TestAccCockpitSources_DataSource_Empty(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_account_project" "project" {
						name = "tf_tests_cockpit_sources_empty"
				  	}

					data "scaleway_cockpit_sources" "empty" {
					  project_id = scaleway_account_project.project.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaleway_cockpit_sources.empty", "sources.#", "0"),
				),
			},
		},
	})
}

func TestAccCockpitSources_DataSource_DefaultSources(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	orgID, orgIDExists := tt.Meta.ScwClient().GetDefaultOrganizationID()
	if !orgIDExists {
		orgID = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					data scaleway_account_project "by_name" {
						name = "default"
						organization_id = "%s"
					}

					data "scaleway_cockpit_sources" "default_sources" {
					  	project_id = data.scaleway_account_project.by_name.id
						origin     = "scaleway"
					}
				`, orgID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaleway_cockpit_sources.default_sources", "sources.#", "2"),
					resource.TestCheckResourceAttr("data.scaleway_cockpit_sources.default_sources", "sources.0.origin", "scaleway"),
					resource.TestCheckResourceAttr("data.scaleway_cockpit_sources.default_sources", "sources.1.origin", "scaleway"),
				),
			},
		},
	})
}
