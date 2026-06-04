package edgeservices_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	edgeservicestestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/edgeservices/testfuncs"
)

func TestAccDataSourceWAFStage_ByID(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             edgeservicestestfuncs.CheckEdgeServicesWAFDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_edge_services_pipeline" "main" {
					  name        = "tf-tests-ds-waf-id"
					  description = "pipeline for WAF data source test"
					}

					resource "scaleway_object_bucket" "main" {
					  name = "test-acc-scaleway-object-bucket-ds-waf-id"
					}

					resource "scaleway_edge_services_backend_stage" "main" {
					  pipeline_id = scaleway_edge_services_pipeline.main.id
					  s3_backend_config {
					    bucket_name   = scaleway_object_bucket.main.name
					    bucket_region = "fr-par"
					  }
					}

					resource "scaleway_edge_services_waf_stage" "main" {
					  pipeline_id      = scaleway_edge_services_pipeline.main.id
					  backend_stage_id = scaleway_edge_services_backend_stage.main.id
					  mode             = "enable"
					  paranoia_level   = 2
					}

					data "scaleway_edge_services_waf_stage" "by_id" {
					  waf_stage_id = scaleway_edge_services_waf_stage.main.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					edgeservicestestfuncs.CheckEdgeServicesWAFExists(tt, "scaleway_edge_services_waf_stage.main"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_edge_services_waf_stage.by_id", "pipeline_id",
						"scaleway_edge_services_waf_stage.main", "pipeline_id"),
					resource.TestCheckResourceAttr("data.scaleway_edge_services_waf_stage.by_id", "paranoia_level", "2"),
					resource.TestCheckResourceAttr("data.scaleway_edge_services_waf_stage.by_id", "mode", "enable"),
					resource.TestCheckResourceAttrSet("data.scaleway_edge_services_waf_stage.by_id", "created_at"),
				),
			},
		},
	})
}

func TestAccDataSourceWAFStage_ByPipelineID(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             edgeservicestestfuncs.CheckEdgeServicesWAFDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_edge_services_pipeline" "main" {
					  name        = "tf-tests-ds-waf-filter"
					  description = "pipeline for WAF filter test"
					}

					resource "scaleway_object_bucket" "main" {
					  name = "test-acc-scaleway-object-bucket-ds-waf-filter"
					}

					resource "scaleway_edge_services_backend_stage" "main" {
					  pipeline_id = scaleway_edge_services_pipeline.main.id
					  s3_backend_config {
					    bucket_name   = scaleway_object_bucket.main.name
					    bucket_region = "fr-par"
					  }
					}

					resource "scaleway_edge_services_waf_stage" "main" {
					  pipeline_id      = scaleway_edge_services_pipeline.main.id
					  backend_stage_id = scaleway_edge_services_backend_stage.main.id
					  mode             = "enable"
					  paranoia_level   = 2
					}

					data "scaleway_edge_services_waf_stage" "by_pipeline" {
					  pipeline_id = scaleway_edge_services_pipeline.main.id
					  depends_on  = [scaleway_edge_services_waf_stage.main]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.scaleway_edge_services_waf_stage.by_pipeline", "pipeline_id",
						"scaleway_edge_services_waf_stage.main", "pipeline_id"),
					resource.TestCheckResourceAttr("data.scaleway_edge_services_waf_stage.by_pipeline", "paranoia_level", "2"),
					resource.TestCheckResourceAttr("data.scaleway_edge_services_waf_stage.by_pipeline", "mode", "enable"),
				),
			},
		},
	})
}
