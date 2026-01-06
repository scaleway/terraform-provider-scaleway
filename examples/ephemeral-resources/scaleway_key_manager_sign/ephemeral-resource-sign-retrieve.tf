# Generate a signature for a digest using a key manager key with rsa_pss_2048_sha256 algorithm
resource "scaleway_key_manager_key" "my_key" {
  name        = "my-key"
  region      = "fr-par"
  usage       = "asymmetric_signing"
  algorithm   = "rsa_pss_2048_sha256"
  unprotected = true
}
ephemeral "scaleway_key_manager_sign" "main" {
  key_id = scaleway_key_manager_key.my_key.id
  digest = "my-base64-digest"
  region = "fr-par"
}
# Create a secret, and store the signature in a secret version.
resource "scaleway_secret" "main" {
  name = "my-secret"
}
resource "scaleway_secret_version" "v1" {
  description = "my-secret-version"
  secret_id   = scaleway_secret.main.id
  data_wo     = ephemeral.scaleway_key_manager_sign.main.signature
}
# Retrieve the signature from the secret version datasource
data "scaleway_secret_version" "data_v1" {
  secret_id  = scaleway_secret.main.id
  revision   = "1"
  depends_on = [scaleway_secret_version.v1]
}
