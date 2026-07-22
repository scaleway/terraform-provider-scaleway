package edgeservices_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	edgeservicestestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/edgeservices/testfuncs"
)

func TestAccDataSourceRouteStage_ByID(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             edgeservicestestfuncs.CheckEdgeServicesRouteDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_edge_services_plan" "main" {
					  name = "starter"
					}

					resource "scaleway_edge_services_pipeline" "main" {
					  name        = "tf-tests-ds-route-id"
					  description = "pipeline for route data source test"
					  depends_on  = [scaleway_edge_services_plan.main]
					}

					resource "scaleway_edge_services_waf_stage" "waf" {
					  pipeline_id    = scaleway_edge_services_pipeline.main.id
					  mode           = "enable"
					  paranoia_level = 3
					}

					resource "scaleway_object_bucket" "main" {
					  name = "test-acc-scaleway-object-bucket-ds-route-id"
					}

					resource "scaleway_edge_services_backend_stage" "backend" {
					  pipeline_id = scaleway_edge_services_pipeline.main.id
					  s3_backend_config {
					    bucket_name   = scaleway_object_bucket.main.name
					    bucket_region = "fr-par"
					  }
					}

					resource "scaleway_edge_services_route_stage" "main" {
					  pipeline_id  = scaleway_edge_services_pipeline.main.id
					  waf_stage_id = scaleway_edge_services_waf_stage.waf.id

					  rule {
					    backend_stage_id = scaleway_edge_services_backend_stage.backend.id
					    rule_http_match {
					      method_filters = ["get"]
					      path_filter {
					        path_filter_type = "regex"
					        value            = ".*"
					      }
					    }
					  }
					}

					data "scaleway_edge_services_route_stage" "by_id" {
					  route_stage_id = scaleway_edge_services_route_stage.main.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					edgeservicestestfuncs.CheckEdgeServicesRouteExists(tt, "scaleway_edge_services_route_stage.main"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_edge_services_route_stage.by_id", "pipeline_id",
						"scaleway_edge_services_route_stage.main", "pipeline_id"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_edge_services_route_stage.by_id", "waf_stage_id",
						"scaleway_edge_services_route_stage.main", "waf_stage_id"),
					resource.TestCheckResourceAttr("data.scaleway_edge_services_route_stage.by_id", "rule.0.rule_http_match.0.method_filters.0", "get"),
					resource.TestCheckResourceAttrSet("data.scaleway_edge_services_route_stage.by_id", "created_at"),
				),
			},
		},
	})
}

func TestAccDataSourceRouteStage_ByPipelineID(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             edgeservicestestfuncs.CheckEdgeServicesRouteDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_edge_services_plan" "main" {
					  name = "starter"
					}

					resource "scaleway_edge_services_pipeline" "main" {
					  name        = "tf-tests-ds-route-filter"
					  description = "pipeline for route filter test"
					  depends_on  = [scaleway_edge_services_plan.main]
					}

					resource "scaleway_edge_services_waf_stage" "waf" {
					  pipeline_id    = scaleway_edge_services_pipeline.main.id
					  mode           = "enable"
					  paranoia_level = 3
					}

					resource "scaleway_object_bucket" "main" {
					  name = "test-acc-scaleway-object-bucket-ds-route-filter"
					}

					resource "scaleway_edge_services_backend_stage" "backend" {
					  pipeline_id = scaleway_edge_services_pipeline.main.id
					  s3_backend_config {
					    bucket_name   = scaleway_object_bucket.main.name
					    bucket_region = "fr-par"
					  }
					}

					resource "scaleway_edge_services_route_stage" "main" {
					  pipeline_id  = scaleway_edge_services_pipeline.main.id
					  waf_stage_id = scaleway_edge_services_waf_stage.waf.id

					  rule {
					    backend_stage_id = scaleway_edge_services_backend_stage.backend.id
					    rule_http_match {
					      method_filters = ["get"]
					      path_filter {
					        path_filter_type = "regex"
					        value            = ".*"
					      }
					    }
					  }
					}

					data "scaleway_edge_services_route_stage" "by_pipeline" {
					  pipeline_id = scaleway_edge_services_pipeline.main.id
					  depends_on  = [scaleway_edge_services_route_stage.main]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.scaleway_edge_services_route_stage.by_pipeline", "pipeline_id",
						"scaleway_edge_services_route_stage.main", "pipeline_id"),
					resource.TestCheckResourceAttr("data.scaleway_edge_services_route_stage.by_pipeline", "rule.0.rule_http_match.0.method_filters.0", "get"),
				),
			},
		},
	})
}
