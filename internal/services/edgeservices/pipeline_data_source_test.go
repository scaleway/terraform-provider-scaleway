package edgeservices_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	edgeservicestestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/edgeservices/testfuncs"
)

func TestAccDataSourcePipeline_ByID(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             edgeservicestestfuncs.CheckEdgeServicesPipelineDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_edge_services_plan" "main" {
					  name = "starter"
					}

					resource "scaleway_edge_services_pipeline" "main" {
					  name        = "tf-tests-ds-pipeline-id"
					  description = "pipeline for data source test"
					  depends_on  = [scaleway_edge_services_plan.main]
					}

					data "scaleway_edge_services_pipeline" "by_id" {
					  pipeline_id = scaleway_edge_services_pipeline.main.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					edgeservicestestfuncs.CheckEdgeServicesPipelineExists(tt, "scaleway_edge_services_pipeline.main"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_edge_services_pipeline.by_id", "name",
						"scaleway_edge_services_pipeline.main", "name"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_edge_services_pipeline.by_id", "description",
						"scaleway_edge_services_pipeline.main", "description"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_edge_services_pipeline.by_id", "project_id",
						"scaleway_edge_services_pipeline.main", "project_id"),
					resource.TestCheckResourceAttrSet("data.scaleway_edge_services_pipeline.by_id", "created_at"),
					resource.TestCheckResourceAttrSet("data.scaleway_edge_services_pipeline.by_id", "updated_at"),
				),
			},
		},
	})
}

func TestAccDataSourcePipeline_ByName(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             edgeservicestestfuncs.CheckEdgeServicesPipelineDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_edge_services_plan" "main" {
					  name = "starter"
					}

					resource "scaleway_edge_services_pipeline" "main" {
					  name        = "tf-tests-ds-pipeline-name"
					  description = "pipeline for name filter test"
					  depends_on  = [scaleway_edge_services_plan.main]
					}

					data "scaleway_edge_services_pipeline" "by_name" {
					  name = scaleway_edge_services_pipeline.main.name
					  depends_on = [scaleway_edge_services_pipeline.main]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					edgeservicestestfuncs.CheckEdgeServicesPipelineExists(tt, "scaleway_edge_services_pipeline.main"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_edge_services_pipeline.by_name", "name",
						"scaleway_edge_services_pipeline.main", "name"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_edge_services_pipeline.by_name", "description",
						"scaleway_edge_services_pipeline.main", "description"),
				),
			},
		},
	})
}
