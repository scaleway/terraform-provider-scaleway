package edgeservices_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	edgeservicestestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/edgeservices/testfuncs"
)

func TestAccDataSourceBackendStage_ByID(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             edgeservicestestfuncs.CheckEdgeServicesBackendDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_edge_services_plan" "main" {
					  name = "starter"
					}

					resource "scaleway_edge_services_pipeline" "main" {
					  name        = "tf-tests-ds-backend-id"
					  description = "pipeline for backend data source test"
					  depends_on  = [scaleway_edge_services_plan.main]
					}

					resource "scaleway_object_bucket" "main" {
					  name = "test-acc-scaleway-object-bucket-ds-backend-id"
					}

					resource "scaleway_edge_services_backend_stage" "main" {
					  pipeline_id = scaleway_edge_services_pipeline.main.id
					  s3_backend_config {
					    bucket_name   = scaleway_object_bucket.main.name
					    bucket_region = "fr-par"
					  }
					}

					data "scaleway_edge_services_backend_stage" "by_id" {
					  backend_stage_id = scaleway_edge_services_backend_stage.main.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					edgeservicestestfuncs.CheckEdgeServicesBackendExists(tt, "scaleway_edge_services_backend_stage.main"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_edge_services_backend_stage.by_id", "pipeline_id",
						"scaleway_edge_services_backend_stage.main", "pipeline_id"),
					resource.TestCheckResourceAttr("data.scaleway_edge_services_backend_stage.by_id", "s3_backend_config.0.bucket_region", "fr-par"),
					resource.TestCheckResourceAttrSet("data.scaleway_edge_services_backend_stage.by_id", "created_at"),
				),
			},
		},
	})
}

func TestAccDataSourceBackendStage_ByPipelineID(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             edgeservicestestfuncs.CheckEdgeServicesBackendDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_edge_services_plan" "main" {
					  name = "starter"
					}

					resource "scaleway_edge_services_pipeline" "main" {
					  name        = "tf-tests-ds-backend-filter"
					  description = "pipeline for backend filter test"
					  depends_on  = [scaleway_edge_services_plan.main]
					}

					resource "scaleway_object_bucket" "main" {
					  name = "test-acc-scaleway-object-bucket-ds-backend-filter"
					}

					resource "scaleway_edge_services_backend_stage" "main" {
					  pipeline_id = scaleway_edge_services_pipeline.main.id
					  s3_backend_config {
					    bucket_name   = scaleway_object_bucket.main.name
					    bucket_region = "fr-par"
					  }
					}

					data "scaleway_edge_services_backend_stage" "by_pipeline" {
					  pipeline_id = scaleway_edge_services_pipeline.main.id
					  depends_on  = [scaleway_edge_services_backend_stage.main]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.scaleway_edge_services_backend_stage.by_pipeline", "pipeline_id",
						"scaleway_edge_services_backend_stage.main", "pipeline_id"),
					resource.TestCheckResourceAttr("data.scaleway_edge_services_backend_stage.by_pipeline", "s3_backend_config.0.bucket_region", "fr-par"),
				),
			},
		},
	})
}
