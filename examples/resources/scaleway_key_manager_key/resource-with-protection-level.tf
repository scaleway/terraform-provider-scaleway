# Create a key with a dedicated protection level.
# In the example below, the key's cryptographic operations are performed within a dedicated Hardware Security Module (HSM).

resource "scaleway_key_manager_key" "hsm" {
  name             = "my-kms-key-hsm"
  usage            = "symmetric_encryption"
  algorithm        = "aes_256_gcm"
  protection_level = "hsm"
  unprotected      = true
}
