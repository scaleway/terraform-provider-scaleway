package edgeservices_test

import (
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
			{
				ResourceName:      "scaleway_edge_services_waf_stage.main",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEdgeServicesWAF_ExclusionRules(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             edgeservicestestfuncs.CheckEdgeServicesWAFDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_edge_services_pipeline" "main" {
					  name        = "my-edge_services-pipeline-waf-exclusions"
					  description = "pipeline description"
					}

					resource "scaleway_edge_services_waf_stage" "main" {
                      pipeline_id    = scaleway_edge_services_pipeline.main.id
					  mode           = "enable"
					  paranoia_level = 3

					  exclusion_rules {
					    rule_id = 942100
					  }

					  exclusion_rules {
					    rule_id = 920350
					  }
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					edgeservicestestfuncs.CheckEdgeServicesWAFExists(tt, "scaleway_edge_services_waf_stage.main"),
					resource.TestCheckResourceAttr("scaleway_edge_services_waf_stage.main", "exclusion_rules.#", "2"),
					resource.TestCheckResourceAttr("scaleway_edge_services_waf_stage.main", "exclusion_rules.0.rule_id", "942100"),
					resource.TestCheckResourceAttr("scaleway_edge_services_waf_stage.main", "exclusion_rules.1.rule_id", "920350"),
				),
			},
			{
				Config: `
					resource "scaleway_edge_services_pipeline" "main" {
					  name        = "my-edge_services-pipeline-waf-exclusions"
					  description = "pipeline description"
					}

					resource "scaleway_edge_services_waf_stage" "main" {
                      pipeline_id    = scaleway_edge_services_pipeline.main.id
					  mode           = "enable"
					  paranoia_level = 3

					  exclusion_rules {
					    rule_id = 941100
					  }
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					edgeservicestestfuncs.CheckEdgeServicesWAFExists(tt, "scaleway_edge_services_waf_stage.main"),
					resource.TestCheckResourceAttr("scaleway_edge_services_waf_stage.main", "exclusion_rules.#", "1"),
					resource.TestCheckResourceAttr("scaleway_edge_services_waf_stage.main", "exclusion_rules.0.rule_id", "941100"),
				),
			},
			{
				ResourceName:      "scaleway_edge_services_waf_stage.main",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
