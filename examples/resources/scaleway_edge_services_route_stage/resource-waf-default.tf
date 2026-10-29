resource "scaleway_edge_services_route_stage" "main" {
  pipeline_id  = scaleway_edge_services_pipeline.main.id
  waf_stage_id = scaleway_edge_services_waf_stage.waf.id

  rule {
    backend_stage_id = scaleway_edge_services_backend_stage.backend.id
    rule_http_match {
      method_filters = ["get", "post"]
      path_filter {
        path_filter_type = "regex"
        value            = ".*"
      }
    }
  }
}
