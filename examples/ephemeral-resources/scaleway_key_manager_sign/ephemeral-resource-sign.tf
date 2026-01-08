# Generate a signature for a digest using a key manager key with ec_p256_sha256 algorithm
resource "scaleway_key_manager_key" "my_key" {
  name        = "my-key"
  region      = "fr-par"
  usage       = "asymmetric_signing"
  algorithm   = "ec_p256_sha256"
  unprotected = true
}
ephemeral "scaleway_key_manager_sign" "main" {
  key_id = scaleway_key_manager_key.my_key.id
  digest = "my-base64-digest"
  region = "fr-par"
}
