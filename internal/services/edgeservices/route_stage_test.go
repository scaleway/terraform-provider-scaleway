package edgeservices_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	edgeservicestestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/edgeservices/testfuncs"
)

func TestAccEdgeServicesRoute_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      edgeservicestestfuncs.CheckEdgeServicesRouteDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
				resource "scaleway_edge_services_pipeline" "main" {
				  name        = "my-edge-services-pipeline"
				  description = "pipeline description"
				}
				
				resource "scaleway_edge_services_waf_stage" "waf" {
				  pipeline_id    = scaleway_edge_services_pipeline.main.id
				  mode           = "enable"
				  paranoia_level = 3
				}
				
				resource "scaleway_object_bucket" "main" {
				  name = "test-acc-scaleway-object-bucket-basic-route"
				  tags = {
					foo = "bar"
				  }
				}

				resource "scaleway_edge_services_backend_stage" "backend" {
				  pipeline_id = scaleway_edge_services_pipeline.main.id
				  s3_backend_config {
					bucket_name   = scaleway_object_bucket.main.name
					bucket_region = "fr-par"
				  }
				}
				
				resource "scaleway_edge_services_route_stage" "main" {
				  pipeline_id   = scaleway_edge_services_pipeline.main.id
				  waf_stage_id  = scaleway_edge_services_waf_stage.waf.id
				
				  rule {
					backend_stage_id = scaleway_edge_services_backend_stage.backend.id
					rule_http_match {
					  method_filters = ["get", "post"]
					  path_filter {
						path_filter_type = "regex"
						value           = ".*"
					  }
					}
				  }
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					edgeservicestestfuncs.CheckEdgeServicesRouteExists(tt, "scaleway_edge_services_route_stage.main"),
					resource.TestCheckResourceAttrPair(
						"scaleway_edge_services_backend_stage.backend", "id",
						"scaleway_edge_services_route_stage.main", "rule.0.backend_stage_id"),
					resource.TestCheckResourceAttrPair(
						"scaleway_edge_services_waf_stage.waf", "id",
						"scaleway_edge_services_route_stage.main", "waf_stage_id"),
					resource.TestCheckResourceAttrPair(
						"scaleway_edge_services_waf_stage.waf", "id",
						"scaleway_edge_services_route_stage.main", "waf_stage_id"),
					resource.TestCheckResourceAttr("scaleway_edge_services_route_stage.main", "rule.0.rule_http_match.0.method_filters.0", "get"),
					resource.TestCheckResourceAttr("scaleway_edge_services_route_stage.main", "rule.0.rule_http_match.0.method_filters.1", "post"),
					resource.TestCheckResourceAttr("scaleway_edge_services_route_stage.main", "rule.0.rule_http_match.0.path_filter.0.path_filter_type", "regex"),
					resource.TestCheckResourceAttr("scaleway_edge_services_route_stage.main", "rule.0.rule_http_match.0.path_filter.0.value", ".*"),
					resource.TestCheckResourceAttrSet("scaleway_edge_services_route_stage.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_edge_services_route_stage.main", "updated_at"),
				),
			},
		},
	})
}
