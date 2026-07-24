resource "scaleway_annotations_key" "environment" {
  name        = "environment"
  description = "Deployment environment (production, staging, development)"
}

data "scaleway_annotations_key" "main" {
  key_id = scaleway_annotations_key.environment.id
}
