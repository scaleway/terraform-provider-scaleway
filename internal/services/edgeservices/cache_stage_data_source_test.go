package edgeservices_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	edgeservicestestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/edgeservices/testfuncs"
)

func TestAccDataSourceCacheStage_ByID(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             edgeservicestestfuncs.CheckEdgeServicesCacheDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_edge_services_pipeline" "main" {
					  name        = "tf-tests-ds-cache-id"
					  description = "pipeline for cache data source test"
					}

					resource "scaleway_object_bucket" "main" {
					  name = "test-acc-scaleway-object-bucket-ds-cache-id"
					}

					resource "scaleway_edge_services_backend_stage" "main" {
					  pipeline_id = scaleway_edge_services_pipeline.main.id
					  s3_backend_config {
					    bucket_name   = scaleway_object_bucket.main.name
					    bucket_region = "fr-par"
					  }
					}

					resource "scaleway_edge_services_cache_stage" "main" {
					  pipeline_id      = scaleway_edge_services_pipeline.main.id
					  backend_stage_id = scaleway_edge_services_backend_stage.main.id
					  fallback_ttl     = 3600
					}

					data "scaleway_edge_services_cache_stage" "by_id" {
					  cache_stage_id = scaleway_edge_services_cache_stage.main.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					edgeservicestestfuncs.CheckEdgeServicesCacheExists(tt, "scaleway_edge_services_cache_stage.main"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_edge_services_cache_stage.by_id", "pipeline_id",
						"scaleway_edge_services_cache_stage.main", "pipeline_id"),
					resource.TestCheckResourceAttr("data.scaleway_edge_services_cache_stage.by_id", "fallback_ttl", "3600"),
					resource.TestCheckResourceAttrSet("data.scaleway_edge_services_cache_stage.by_id", "created_at"),
				),
			},
		},
	})
}

func TestAccDataSourceCacheStage_ByPipelineID(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             edgeservicestestfuncs.CheckEdgeServicesCacheDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_edge_services_pipeline" "main" {
					  name        = "tf-tests-ds-cache-filter"
					  description = "pipeline for cache filter test"
					}

					resource "scaleway_object_bucket" "main" {
					  name = "test-acc-scaleway-object-bucket-ds-cache-filter"
					}

					resource "scaleway_edge_services_backend_stage" "main" {
					  pipeline_id = scaleway_edge_services_pipeline.main.id
					  s3_backend_config {
					    bucket_name   = scaleway_object_bucket.main.name
					    bucket_region = "fr-par"
					  }
					}

					resource "scaleway_edge_services_cache_stage" "main" {
					  pipeline_id      = scaleway_edge_services_pipeline.main.id
					  backend_stage_id = scaleway_edge_services_backend_stage.main.id
					  fallback_ttl     = 7200
					}

					data "scaleway_edge_services_cache_stage" "by_pipeline" {
					  pipeline_id = scaleway_edge_services_pipeline.main.id
					  depends_on  = [scaleway_edge_services_cache_stage.main]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.scaleway_edge_services_cache_stage.by_pipeline", "pipeline_id",
						"scaleway_edge_services_cache_stage.main", "pipeline_id"),
					resource.TestCheckResourceAttr("data.scaleway_edge_services_cache_stage.by_pipeline", "fallback_ttl", "7200"),
				),
			},
		},
	})
}
