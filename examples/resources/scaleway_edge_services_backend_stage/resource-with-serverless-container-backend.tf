### With Serverless Container backend

resource "scaleway_container_namespace" "main" {
  name = "my-namespace"
}

resource "scaleway_container" "main" {
  namespace_id = scaleway_container_namespace.main.id
  name         = "my-container"
  image        = "nginx:1.29.4-alpine"
  port         = 80
}

resource "scaleway_edge_services_pipeline" "main" {
  name = "my-pipeline"
}

resource "scaleway_edge_services_backend_stage" "main" {
  pipeline_id = scaleway_edge_services_pipeline.main.id
  container_backend_config {
    container_id = scaleway_container.main.id
    region       = "fr-par"
  }
}
