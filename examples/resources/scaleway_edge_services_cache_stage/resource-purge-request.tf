### Purge request

resource "scaleway_edge_services_cache_stage" "main" {
  pipeline_id      = scaleway_edge_services_pipeline.main.id
  backend_stage_id = scaleway_edge_services_backend_stage.main.id

  purge {
    pipeline_id = scaleway_edge_services_pipeline.main.id
    all         = true
  }
}
