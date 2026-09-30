### With LB backend

resource "scaleway_lb" "main" {
  ip_ids = [scaleway_lb_ip.main.id]
  zone   = "fr-par-1"
  type   = "LB-S"
}

resource "scaleway_lb_frontend" "main" {
  lb_id        = scaleway_lb.main.id
  backend_id   = scaleway_lb_backend.main.id
  name         = "frontend01"
  inbound_port = "443"
  certificate_ids = [
    scaleway_lb_certificate.cert01.id,
  ]
}

resource "scaleway_edge_services_pipeline" "main" {
  name = "my-pipeline"
}

resource "scaleway_edge_services_backend_stage" "main" {
  pipeline_id = scaleway_edge_services_pipeline.main.id
  lb_backend_config {
    lb_config {
      id          = scaleway_lb.main.id
      frontend_id = scaleway_lb_frontend.id
      is_ssl      = true
      zone        = "fr-par-1"
    }
  }
}
