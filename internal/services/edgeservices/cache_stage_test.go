package edgeservices_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	edgeservicestestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/edgeservices/testfuncs"
)

func TestAccEdgeServicesCache_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             edgeservicestestfuncs.CheckEdgeServicesCacheDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_edge_services_pipeline" "main" {
					  name        = "my-edge_services-pipeline"
					  description = "pipeline description"
					}

					resource "scaleway_edge_services_cache_stage" "main" {
                      pipeline_id = scaleway_edge_services_pipeline.main.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					edgeservicestestfuncs.CheckEdgeServicesCacheExists(tt, "scaleway_edge_services_cache_stage.main"),
					resource.TestCheckResourceAttrPair(
						"scaleway_edge_services_pipeline.main", "id",
						"scaleway_edge_services_cache_stage.main", "pipeline_id"),
					resource.TestCheckResourceAttr("scaleway_edge_services_cache_stage.main", "fallback_ttl", "3600"),
					resource.TestCheckResourceAttr("scaleway_edge_services_cache_stage.main", "include_cookies", "false"),
					resource.TestCheckResourceAttrSet("scaleway_edge_services_cache_stage.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_edge_services_cache_stage.main", "updated_at"),
				),
			},
			{
				Config: `
					resource "scaleway_edge_services_pipeline" "main" {
					  name        = "my-edge_services-pipeline"
					  description = "pipeline description"
					}

					resource "scaleway_edge_services_cache_stage" "main" {
                      pipeline_id     = scaleway_edge_services_pipeline.main.id
					  include_cookies = true
                      fallback_ttl    = 7200
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					edgeservicestestfuncs.CheckEdgeServicesCacheExists(tt, "scaleway_edge_services_cache_stage.main"),
					resource.TestCheckResourceAttrPair(
						"scaleway_edge_services_pipeline.main", "id",
						"scaleway_edge_services_cache_stage.main", "pipeline_id"),
					resource.TestCheckResourceAttr("scaleway_edge_services_cache_stage.main", "fallback_ttl", "7200"),
					resource.TestCheckResourceAttr("scaleway_edge_services_cache_stage.main", "include_cookies", "true"),
					resource.TestCheckResourceAttrSet("scaleway_edge_services_cache_stage.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_edge_services_cache_stage.main", "updated_at"),
				),
			},
		},
	})
}
