### Basic

resource "scaleway_edge_services_waf_stage" "main" {
  pipeline_id    = scaleway_edge_services_pipeline.main.id
  mode           = "enable"
  paranoia_level = 3
}
