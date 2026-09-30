### Managed

resource "scaleway_edge_services_tls_stage" "main" {
  pipeline_id         = scaleway_edge_services_pipeline.main.id
  managed_certificate = true
}
