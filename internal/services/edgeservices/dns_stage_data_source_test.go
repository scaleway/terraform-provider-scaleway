package edgeservices_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	edgeservicestestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/edgeservices/testfuncs"
)

func TestAccDataSourceDNSStage_ByID(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             edgeservicestestfuncs.CheckEdgeServicesDNSDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_edge_services_plan" "main" {
					  name = "starter"
					}

					resource "scaleway_edge_services_pipeline" "main" {
					  name        = "tf-tests-ds-dns-id"
					  description = "pipeline for DNS data source test"
					  depends_on  = [scaleway_edge_services_plan.main]
					}

					resource "scaleway_object_bucket" "main" {
					  name = "test-acc-scaleway-object-bucket-ds-dns-id"
					}

					resource "scaleway_edge_services_backend_stage" "main" {
					  pipeline_id = scaleway_edge_services_pipeline.main.id
					  s3_backend_config {
					    bucket_name   = scaleway_object_bucket.main.name
					    bucket_region = "fr-par"
					  }
					}

					resource "scaleway_edge_services_dns_stage" "main" {
					  pipeline_id      = scaleway_edge_services_pipeline.main.id
					  backend_stage_id = scaleway_edge_services_backend_stage.main.id
					}

					data "scaleway_edge_services_dns_stage" "by_id" {
					  dns_stage_id = scaleway_edge_services_dns_stage.main.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					edgeservicestestfuncs.CheckEdgeServicesDNSExists(tt, "scaleway_edge_services_dns_stage.main"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_edge_services_dns_stage.by_id", "pipeline_id",
						"scaleway_edge_services_dns_stage.main", "pipeline_id"),
					resource.TestCheckResourceAttrSet("data.scaleway_edge_services_dns_stage.by_id", "created_at"),
				),
			},
		},
	})
}

func TestAccDataSourceDNSStage_ByPipelineID(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             edgeservicestestfuncs.CheckEdgeServicesDNSDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_edge_services_plan" "main" {
					  name = "starter"
					}

					resource "scaleway_edge_services_pipeline" "main" {
					  name        = "tf-tests-ds-dns-filter"
					  description = "pipeline for DNS filter test"
					  depends_on  = [scaleway_edge_services_plan.main]
					}

					resource "scaleway_object_bucket" "main" {
					  name = "test-acc-scaleway-object-bucket-ds-dns-filter"
					}

					resource "scaleway_edge_services_backend_stage" "main" {
					  pipeline_id = scaleway_edge_services_pipeline.main.id
					  s3_backend_config {
					    bucket_name   = scaleway_object_bucket.main.name
					    bucket_region = "fr-par"
					  }
					}

					resource "scaleway_edge_services_dns_stage" "main" {
					  pipeline_id      = scaleway_edge_services_pipeline.main.id
					  backend_stage_id = scaleway_edge_services_backend_stage.main.id
					}

					data "scaleway_edge_services_dns_stage" "by_pipeline" {
					  pipeline_id = scaleway_edge_services_pipeline.main.id
					  depends_on  = [scaleway_edge_services_dns_stage.main]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.scaleway_edge_services_dns_stage.by_pipeline", "pipeline_id",
						"scaleway_edge_services_dns_stage.main", "pipeline_id"),
				),
			},
		},
	})
}
