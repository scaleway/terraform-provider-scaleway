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
			{
				ResourceName:      "scaleway_edge_services_dns_stage.main",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEdgeServicesDNS_Wildcard(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             edgeservicestestfuncs.CheckEdgeServicesDNSDestroy(tt),
		Steps: []resource.TestStep{
			{
				// A wildcard DNS stage requires the linked TLS stage to hold a valid, CA-signed
				// wildcard certificate. Self-signed certificates are rejected by Edge Services.
				// The certificate is provided through a pre-existing secret created manually in the recording project.
				Config: `
					data "scaleway_secret" "main" {
					  secret_id = "d023fa93-eb36-44b5-91b8-13b998e2e630"
					}
	
					resource "scaleway_edge_services_pipeline" "main" {
					  name        = "tf-tests-edge-dns-wildcard"
					  description = "pipeline for wildcard DNS test"
					}

					resource "scaleway_object_bucket" "main" {
					  name = "test-acc-edge-dns-wildcard"
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
					  managed_certificate = false
					  secrets {
					    secret_id = data.scaleway_secret.main.secret_id
					    region    = "fr-par"
					  }
					}

					resource "scaleway_edge_services_dns_stage" "main" {
					  pipeline_id     = scaleway_edge_services_pipeline.main.id
					  tls_stage_id    = scaleway_edge_services_tls_stage.main.id
					  fqdns           = ["edge-dns-wc.tf.scaleway-terraform.com"]
					  wildcard_domain = true
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					edgeservicestestfuncs.CheckEdgeServicesDNSExists(tt, "scaleway_edge_services_dns_stage.main"),
					resource.TestCheckResourceAttrPair(
						"scaleway_edge_services_pipeline.main", "id",
						"scaleway_edge_services_dns_stage.main", "pipeline_id"),
					resource.TestCheckResourceAttrPair(
						"scaleway_edge_services_tls_stage.main", "id",
						"scaleway_edge_services_dns_stage.main", "tls_stage_id"),
					resource.TestCheckResourceAttr("scaleway_edge_services_dns_stage.main", "wildcard_domain", "true"),
					resource.TestCheckResourceAttr("scaleway_edge_services_dns_stage.main", "fqdns.0", "edge-dns-wc.tf.scaleway-terraform.com"),
					resource.TestCheckResourceAttrSet("scaleway_edge_services_dns_stage.main", "type"),
					resource.TestCheckResourceAttrSet("scaleway_edge_services_dns_stage.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_edge_services_dns_stage.main", "updated_at"),
				),
			},
		},
	})
}
