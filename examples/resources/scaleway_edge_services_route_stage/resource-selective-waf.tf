resource "scaleway_edge_services_pipeline" "main" {
  name        = "my-pipeline"
  description = "Static site with WAF-protected API"
}

resource "scaleway_object_bucket" "main" {
  name = "my-static-site"
}

resource "scaleway_edge_services_backend_stage" "static" {
  pipeline_id = scaleway_edge_services_pipeline.main.id
  s3_backend_config {
    bucket_name   = scaleway_object_bucket.main.name
    bucket_region = "fr-par"
  }
}

resource "scaleway_edge_services_waf_stage" "api" {
  pipeline_id      = scaleway_edge_services_pipeline.main.id
  backend_stage_id = scaleway_edge_services_backend_stage.static.id
  mode             = "enable"
  paranoia_level   = 2
}

resource "scaleway_edge_services_route_stage" "main" {
  pipeline_id      = scaleway_edge_services_pipeline.main.id
  backend_stage_id = scaleway_edge_services_backend_stage.static.id

  rule {
    waf_stage_id = scaleway_edge_services_waf_stage.api.id
    rule_http_match {
      method_filters = ["get", "post", "put", "patch", "delete"]
      path_filter {
        path_filter_type = "regex"
        value            = "/api/.*"
      }
    }
  }
}
