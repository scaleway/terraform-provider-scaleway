### With object backend

resource "scaleway_object_bucket" "main" {
  name = "my-bucket-name"
  tags = {
    foo = "bar"
  }
}

resource "scaleway_edge_services_pipeline" "main" {
  name = "my-pipeline"
}

resource "scaleway_edge_services_backend_stage" "main" {
  pipeline_id = scaleway_edge_services_pipeline.main.id
  s3_backend_config {
    bucket_name   = scaleway_object_bucket.main.name
    bucket_region = "fr-par"
  }
}
