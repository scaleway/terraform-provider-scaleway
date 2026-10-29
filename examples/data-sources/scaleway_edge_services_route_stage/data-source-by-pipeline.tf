# Retrieve an Edge Services route stage by pipeline ID
data "scaleway_edge_services_route_stage" "by_pipeline" {
  pipeline_id = scaleway_edge_services_pipeline.main.id
}
