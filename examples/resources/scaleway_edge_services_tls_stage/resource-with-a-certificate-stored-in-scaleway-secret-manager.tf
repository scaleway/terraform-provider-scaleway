### With a certificate stored in Scaleway Secret Manager

resource "scaleway_edge_services_tls_stage" "main" {
  pipeline_id = scaleway_edge_services_pipeline.main.id
  secrets {
    secret_id = "11111111-1111-1111-1111-111111111111"
    region    = "fr-par"
  }
}
