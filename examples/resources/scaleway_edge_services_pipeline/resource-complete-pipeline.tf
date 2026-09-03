### Complete pipeline

resource "scaleway_edge_services_pipeline" "main" {
  name        = "pipeline-name"
  description = "pipeline description"
}

resource "scaleway_edge_services_dns_stage" "main" {
  pipeline_id  = scaleway_edge_services_pipeline.main.id
  tls_stage_id = scaleway_edge_services_tls_stage.main.id
  fqdns        = ["subdomain.example.com"]
}

resource "scaleway_edge_services_tls_stage" "main" {
  pipeline_id         = scaleway_edge_services_pipeline.main.id
  cache_stage_id      = scaleway_edge_services_cache_stage.main.id
  managed_certificate = true
}

resource "scaleway_edge_services_cache_stage" "main" {
  pipeline_id    = scaleway_edge_services_pipeline.main.id
  route_stage_id = scaleway_edge_services_route_stage.main.id
}

resource "scaleway_edge_services_route_stage" "main" {
  pipeline_id  = scaleway_edge_services_pipeline.main.id
  waf_stage_id = scaleway_edge_services_waf_stage.main.id

  rule {
    backend_stage_id = scaleway_edge_services_backend_stage.main.id
    rule_http_match {
      method_filters = ["get", "post"]
      path_filter {
        path_filter_type = "regex"
        value            = ".*"
      }
    }
  }
}

resource "scaleway_edge_services_waf_stage" "main" {
  pipeline_id      = scaleway_edge_services_pipeline.main.id
  backend_stage_id = scaleway_edge_services_backend_stage.main.id
  mode             = "enable"
  paranoia_level   = 3
}

resource "scaleway_edge_services_backend_stage" "main" {
  pipeline_id = scaleway_edge_services_pipeline.main.id
  s3_backend_config {
    bucket_name   = "my-bucket-name"
    bucket_region = "fr-par"
  }
}

resource "scaleway_edge_services_head_stage" "main" {
  pipeline_id   = scaleway_edge_services_pipeline.main.id
  head_stage_id = scaleway_edge_services_dns_stage.main.id
}
