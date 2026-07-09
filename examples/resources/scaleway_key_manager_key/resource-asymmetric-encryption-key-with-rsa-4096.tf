### Asymmetric Encryption Key with RSA-4096

resource "scaleway_key_manager_key" "rsa_4096" {
  name        = "rsa-4096-key"
  region      = "fr-par"
  usage       = "asymmetric_encryption"
  algorithm   = "rsa_oaep_4096_sha256"
  description = "Key for encrypting large files with RSA-4096"
  unprotected = true
}
