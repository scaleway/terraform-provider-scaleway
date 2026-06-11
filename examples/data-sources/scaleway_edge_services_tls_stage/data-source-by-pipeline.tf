# Retrieve an Edge Services TLS stage by pipeline ID
data "scaleway_edge_services_tls_stage" "by_pipeline" {
  pipeline_id = scaleway_edge_services_pipeline.main.id
}
