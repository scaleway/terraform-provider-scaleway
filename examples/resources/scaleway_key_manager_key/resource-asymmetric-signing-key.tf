### Asymmetric Signing Key

resource "scaleway_key_manager_key" "signing" {
  name        = "signing-key"
  region      = "fr-par"
  usage       = "asymmetric_signing"
  algorithm   = "rsa_pss_2048_sha256"
  description = "Key for signing documents"
  unprotected = true
}
