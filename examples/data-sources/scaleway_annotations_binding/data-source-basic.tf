resource "scaleway_annotations_key" "environment" {
  name        = "environment"
  description = "Deployment environment (production, staging, development)"
}

resource "scaleway_annotations_value" "production" {
  key_id      = scaleway_annotations_key.environment.id
  name        = "production"
  description = "Production environment"
}

resource "scaleway_key_manager_key" "main" {
  name        = "example-key"
  region      = "fr-par"
  usage       = "symmetric_encryption"
  algorithm   = "aes_256_gcm"
  description = "Example key for binding"
  unprotected = true
}

resource "scaleway_annotations_binding" "main" {
  srn        = scaleway_key_manager_key.main.srn
  value_id   = scaleway_annotations_value.production.id
}

data "scaleway_annotations_binding" "main" {
  id = scaleway_annotations_binding.main.id
}
