resource "scaleway_key_manager_key" "test_key" {
  name        = "tf-test-generate-data-key"
  region      = "fr-par"
  usage       = "symmetric_encryption"
  algorithm   = "aes_256_gcm"
  unprotected = true
}

ephemeral "scaleway_key_manager_generate_data_key" "main" {
  key_id = scaleway_key_manager_key.test_key.id
  region = "fr-par"
}
