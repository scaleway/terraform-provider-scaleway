package edgeservices_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	edgeservicestestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/edgeservices/testfuncs"
)

func TestAccEdgeServicesDNS_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             edgeservicestestfuncs.CheckEdgeServicesDNSDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_edge_services_pipeline" "main" {
					  name        = "my-edge_services-pipeline"
					  description = "pipeline description"
					}

					resource "scaleway_edge_services_dns_stage" "main" {
                      pipeline_id = scaleway_edge_services_pipeline.main.id
					  fqdns       = ["subodomain.example.fr"]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					edgeservicestestfuncs.CheckEdgeServicesDNSExists(tt, "scaleway_edge_services_dns_stage.main"),
					resource.TestCheckResourceAttrPair(
						"scaleway_edge_services_pipeline.main", "id",
						"scaleway_edge_services_dns_stage.main", "pipeline_id"),
					resource.TestCheckResourceAttr("scaleway_edge_services_dns_stage.main", "fqdns.0", "subodomain.example.fr"),
					resource.TestCheckResourceAttrSet("scaleway_edge_services_dns_stage.main", "type"),
					resource.TestCheckResourceAttrSet("scaleway_edge_services_dns_stage.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_edge_services_dns_stage.main", "updated_at"),
				),
			},
		},
	})
}
