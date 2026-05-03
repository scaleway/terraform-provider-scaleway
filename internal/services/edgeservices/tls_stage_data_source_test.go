package edgeservices_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	edgeservicestestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/edgeservices/testfuncs"
)

func TestAccDataSourceTLSStage_ByID(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             edgeservicestestfuncs.CheckEdgeServicesTLSDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_edge_services_plan" "main" {
					  name = "starter"
					}

					resource "scaleway_edge_services_pipeline" "main" {
					  name        = "tf-tests-ds-tls-id"
					  description = "pipeline for TLS data source test"
					  depends_on  = [scaleway_edge_services_plan.main]
					}

					resource "scaleway_object_bucket" "main" {
					  name = "test-acc-scaleway-object-bucket-ds-tls-id"
					}

					resource "scaleway_edge_services_backend_stage" "main" {
					  pipeline_id = scaleway_edge_services_pipeline.main.id
					  s3_backend_config {
					    bucket_name   = scaleway_object_bucket.main.name
					    bucket_region = "fr-par"
					  }
					}

					resource "scaleway_edge_services_tls_stage" "main" {
					  pipeline_id         = scaleway_edge_services_pipeline.main.id
					  backend_stage_id    = scaleway_edge_services_backend_stage.main.id
					  managed_certificate = true
					}

					data "scaleway_edge_services_tls_stage" "by_id" {
					  tls_stage_id = scaleway_edge_services_tls_stage.main.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					edgeservicestestfuncs.CheckEdgeServicesTLSExists(tt, "scaleway_edge_services_tls_stage.main"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_edge_services_tls_stage.by_id", "pipeline_id",
						"scaleway_edge_services_tls_stage.main", "pipeline_id"),
					resource.TestCheckResourceAttr("data.scaleway_edge_services_tls_stage.by_id", "managed_certificate", "true"),
					resource.TestCheckResourceAttrSet("data.scaleway_edge_services_tls_stage.by_id", "created_at"),
				),
			},
		},
	})
}

func TestAccDataSourceTLSStage_ByPipelineID(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             edgeservicestestfuncs.CheckEdgeServicesTLSDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_edge_services_plan" "main" {
					  name = "starter"
					}

					resource "scaleway_edge_services_pipeline" "main" {
					  name        = "tf-tests-ds-tls-filter"
					  description = "pipeline for TLS filter test"
					  depends_on  = [scaleway_edge_services_plan.main]
					}

					resource "scaleway_object_bucket" "main" {
					  name = "test-acc-scaleway-object-bucket-ds-tls-filter"
					}

					resource "scaleway_edge_services_backend_stage" "main" {
					  pipeline_id = scaleway_edge_services_pipeline.main.id
					  s3_backend_config {
					    bucket_name   = scaleway_object_bucket.main.name
					    bucket_region = "fr-par"
					  }
					}

					resource "scaleway_edge_services_tls_stage" "main" {
					  pipeline_id         = scaleway_edge_services_pipeline.main.id
					  backend_stage_id    = scaleway_edge_services_backend_stage.main.id
					  managed_certificate = true
					}

					data "scaleway_edge_services_tls_stage" "by_pipeline" {
					  pipeline_id = scaleway_edge_services_pipeline.main.id
					  depends_on  = [scaleway_edge_services_tls_stage.main]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.scaleway_edge_services_tls_stage.by_pipeline", "pipeline_id",
						"scaleway_edge_services_tls_stage.main", "pipeline_id"),
					resource.TestCheckResourceAttr("data.scaleway_edge_services_tls_stage.by_pipeline", "managed_certificate", "true"),
				),
			},
		},
	})
}
