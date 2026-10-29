# Retrieve an Edge Services backend stage by pipeline ID
data "scaleway_edge_services_backend_stage" "by_pipeline" {
  pipeline_id = scaleway_edge_services_pipeline.main.id
}
