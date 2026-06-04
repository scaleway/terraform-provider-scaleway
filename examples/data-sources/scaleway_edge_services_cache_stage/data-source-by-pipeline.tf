# Retrieve an Edge Services cache stage by pipeline ID
data "scaleway_edge_services_cache_stage" "by_pipeline" {
  pipeline_id = scaleway_edge_services_pipeline.main.id
}
