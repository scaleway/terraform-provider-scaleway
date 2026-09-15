### With Serverless Function backend

resource "scaleway_function_namespace" "main" {
  name = "my-namespace"
}

resource "scaleway_function" "main" {
  namespace_id = scaleway_function_namespace.main.id
  name         = "my-function"
  runtime      = "node20"
  privacy      = "private"
  handler      = "handler.handle"
}

resource "scaleway_edge_services_pipeline" "main" {
  name = "my-pipeline"
}

resource "scaleway_edge_services_backend_stage" "main" {
  pipeline_id = scaleway_edge_services_pipeline.main.id
  function_backend_config {
    function_id = scaleway_function.main.id
    region      = "fr-par"
  }
}
