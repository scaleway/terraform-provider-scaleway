resource "scaleway_key_manager_key" "test_key" {
  name        = "tf-test-encrypt-key"
  region      = "fr-par"
  usage       = "symmetric_encryption"
  algorithm   = "aes_256_gcm"
  unprotected = true
}

ephemeral "scaleway_key_manager_encrypt" "test_encrypt" {
  key_id    = scaleway_key_manager_key.test_key.id
  plaintext = "test plaintext data"
  region    = "fr-par"
}
