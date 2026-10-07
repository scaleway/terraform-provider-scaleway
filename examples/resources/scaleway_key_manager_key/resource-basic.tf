# Create a basic key

resource "scaleway_key_manager_key" "main" {
  name        = "my-kms-key"
  usage       = "symmetric_encryption"
  algorithm   = "aes_256_gcm"
  unprotected = true
}
