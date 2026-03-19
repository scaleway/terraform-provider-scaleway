# Retrieve an Edge Services WAF stage by pipeline ID
data "scaleway_edge_services_waf_stage" "by_pipeline" {
  pipeline_id = scaleway_edge_services_pipeline.main.id
}
