resource "scaleway_edge_services_pipeline" "main" {
  name        = "my-pipeline"
  description = "Multi-host pipeline with host-based routing"
}

resource "scaleway_object_bucket" "api" {
  name = "my-api-bucket"
}

resource "scaleway_object_bucket" "static" {
  name = "my-static-site"
}

resource "scaleway_edge_services_backend_stage" "api" {
  pipeline_id = scaleway_edge_services_pipeline.main.id
  s3_backend_config {
    bucket_name   = scaleway_object_bucket.api.name
    bucket_region = "fr-par"
  }
}

resource "scaleway_edge_services_backend_stage" "static" {
  pipeline_id = scaleway_edge_services_pipeline.main.id
  s3_backend_config {
    bucket_name   = scaleway_object_bucket.static.name
    bucket_region = "fr-par"
  }
}

resource "scaleway_edge_services_route_stage" "main" {
  pipeline_id      = scaleway_edge_services_pipeline.main.id
  backend_stage_id = scaleway_edge_services_backend_stage.static.id

  rule {
    backend_stage_id = scaleway_edge_services_backend_stage.api.id
    rule_http_match {
      host_filter {
        host_filter_type = "regex"
        value            = "api\\.example\\.com"
      }
    }
  }
}
