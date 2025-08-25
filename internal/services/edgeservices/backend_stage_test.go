package edgeservices_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	edgeservicestestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/edgeservices/testfuncs"
)

func TestAccEdgeServicesBackend_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      edgeservicestestfuncs.CheckEdgeServicesBackendDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_edge_services_pipeline" "main" {
					  name        = "my-edge_services-pipeline"
					  description = "pipeline description"
					}

					resource "scaleway_object_bucket" "main" {
					  name = "test-acc-scaleway-object-bucket-basic-es"
					  tags = {
						foo = "bar"
					  }
					}

					resource "scaleway_edge_services_backend_stage" "main" {
                      pipeline_id = scaleway_edge_services_pipeline.main.id
					  s3_backend_config {
						bucket_name   = scaleway_object_bucket.main.name
						bucket_region = "fr-par"
					  }
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					edgeservicestestfuncs.CheckEdgeServicesBackendExists(tt, "scaleway_edge_services_backend_stage.main"),
					resource.TestCheckResourceAttrPair(
						"scaleway_edge_services_pipeline.main", "id",
						"scaleway_edge_services_backend_stage.main", "pipeline_id"),
					resource.TestCheckResourceAttr("scaleway_edge_services_backend_stage.main", "s3_backend_config.0.is_website", "false"),
					resource.TestCheckResourceAttr("scaleway_edge_services_backend_stage.main", "s3_backend_config.0.bucket_region", "fr-par"),
					resource.TestCheckResourceAttrSet("scaleway_edge_services_backend_stage.main", "s3_backend_config.0.bucket_name"),
					resource.TestCheckResourceAttrSet("scaleway_edge_services_backend_stage.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_edge_services_backend_stage.main", "updated_at"),
				),
			},
		},
	})
}

func TestAccEdgeServicesBackend_LB(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      edgeservicestestfuncs.CheckEdgeServicesBackendDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_lb_ip" "lb_ip" {}
					
					resource "scaleway_lb" "lb01" {
					  name  = "lb_name"
					  ip_id = scaleway_lb_ip.lb_ip.id
					  type  = "LB-S"
					}
					
					resource "scaleway_lb_backend" "bck01" {
					  lb_id            = scaleway_lb.lb01.id
					  name             = "backend"
					  forward_protocol = "http"
					  forward_port     = 80
					  ssl_bridging     = true
					
					  health_check_http {
						uri    = "/healthcheck"
						method = "GET"
						code   = 200
					  }
					}
					
					resource "scaleway_lb_frontend" "frt01" {
					  lb_id        = scaleway_lb.lb01.id
					  backend_id   = scaleway_lb_backend.bck01.id
					  name         = "frontend"
					  inbound_port = "443"
					  certificate_ids = [
						scaleway_lb_certificate.cert01.id,
					  ]
					}
					
					resource "scaleway_lb_certificate" "cert01" {
					  lb_id = scaleway_lb.lb01.id
					  name  = "test-cert"
					  letsencrypt {
						common_name = "${replace(scaleway_lb_ip.lb_ip.ip_address, ".", "-")}.lb.${scaleway_lb.lb01.region}.scw.cloud"
					  }
					}

					resource "scaleway_edge_services_pipeline" "main" {
					  name        = "my-edge_services-pipeline"
					  description = "pipeline description"
					}

					resource "scaleway_edge_services_backend_stage" "main" {
                      pipeline_id = scaleway_edge_services_pipeline.main.id
					  lb_backend_config {
					    lb_config {
						  id          = scaleway_lb.lb01.id
						  frontend_id = scaleway_lb_frontend.frt01.id
						  is_ssl      = true
					    }
					  }
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					edgeservicestestfuncs.CheckEdgeServicesBackendExists(tt, "scaleway_edge_services_backend_stage.main"),
					resource.TestCheckResourceAttrPair(
						"scaleway_edge_services_backend_stage.main", "lb_backend_config.0.lb_config.0.id",
						"scaleway_lb.lb01", "id"),
					resource.TestCheckResourceAttrPair(
						"scaleway_edge_services_backend_stage.main", "lb_backend_config.0.lb_config.0.frontend_id",
						"scaleway_lb_frontend.frt01", "id"),
					resource.TestCheckResourceAttrPair(
						"scaleway_edge_services_pipeline.main", "id",
						"scaleway_edge_services_backend_stage.main", "pipeline_id"),
					resource.TestCheckResourceAttr("scaleway_edge_services_backend_stage.main", "lb_backend_config.0.lb_config.0.is_ssl", "true"),
					resource.TestCheckResourceAttr("scaleway_edge_services_backend_stage.main", "lb_backend_config.0.lb_config.0.zone", "fr-par-1"),
					resource.TestCheckResourceAttrSet("scaleway_edge_services_backend_stage.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_edge_services_backend_stage.main", "updated_at"),
				),
			},
		},
	})
}
