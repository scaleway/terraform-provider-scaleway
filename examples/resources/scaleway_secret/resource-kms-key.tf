# Create a secret with your own Key Manager encryption key

resource "scaleway_key_manager_key" "main" {
  name        = "my-kms-key"
  usage       = "symmetric_encryption"
  algorithm   = "aes_256_gcm"
  unprotected = true
}

resource "scaleway_secret" "kms" {
  name   = "foo"
  key_id = scaleway_key_manager_key.main.id
}
