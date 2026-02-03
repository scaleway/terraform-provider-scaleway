package edgeservices_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	edgeservicestestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/edgeservices/testfuncs"
)

func TestAccEdgeServicesWAF_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             edgeservicestestfuncs.CheckEdgeServicesWAFDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_edge_services_pipeline" "main" {
					  name        = "my-edge_services-pipeline"
					  description = "pipeline description"
					}

					resource "scaleway_edge_services_waf_stage" "main" {
                      pipeline_id    = scaleway_edge_services_pipeline.main.id
					  mode           = "enable"
					  paranoia_level = 3
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					edgeservicestestfuncs.CheckEdgeServicesWAFExists(tt, "scaleway_edge_services_waf_stage.main"),
					resource.TestCheckResourceAttrPair(
						"scaleway_edge_services_pipeline.main", "id",
						"scaleway_edge_services_waf_stage.main", "pipeline_id"),
					resource.TestCheckResourceAttr("scaleway_edge_services_waf_stage.main", "mode", "enable"),
					resource.TestCheckResourceAttr("scaleway_edge_services_waf_stage.main", "paranoia_level", "3"),
					resource.TestCheckResourceAttrSet("scaleway_edge_services_waf_stage.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_edge_services_waf_stage.main", "updated_at"),
				),
			},
		},
	})
}

func TestAccEdgeServicesWAF_RejectS3Backend(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             edgeservicestestfuncs.CheckEdgeServicesWAFDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_object_bucket" "main" {
						name   = "test-waf-s3-backend-rejection"
						region = "fr-par"
					}

					resource "scaleway_edge_services_pipeline" "main" {
						name = "test-waf-s3-pipeline-rejection"
					}

					resource "scaleway_edge_services_backend_stage" "main" {
						pipeline_id = scaleway_edge_services_pipeline.main.id
						s3_backend_config {
							bucket_name   = scaleway_object_bucket.main.name
							bucket_region = "fr-par"
							is_website    = true
						}
					}

					resource "scaleway_edge_services_waf_stage" "main" {
						pipeline_id      = scaleway_edge_services_pipeline.main.id
						backend_stage_id = scaleway_edge_services_backend_stage.main.id
						mode             = "enable"
						paranoia_level   = 3
					}
				`,
				ExpectError: regexp.MustCompile("WAF stage is only supported with Load Balancer backends"),
			},
		},
	})
}
