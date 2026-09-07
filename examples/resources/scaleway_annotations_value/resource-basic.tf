resource "scaleway_annotations_key" "environment" {
  name        = "environment"
  description = "Deployment environment (production, staging, development)"
}

resource "scaleway_annotations_value" "production" {
  key_id      = scaleway_annotations_key.environment.id
  name        = "production"
  description = "Production environment"
}
