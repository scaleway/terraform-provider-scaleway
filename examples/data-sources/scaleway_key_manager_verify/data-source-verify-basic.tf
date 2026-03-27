# The following commands allow you to:

# - create a key named `my-kms-key`
# - generate a signature for a message using the key created above
# - store the signature in a secret manager secret version
# - verify the signature using the key created above, the digest and the signature retrieved from the secret

// Create a key
resource "scaleway_key_manager_key" "main" {
  name        = "my-kms-key"
  region      = "fr-par"
  usage       = "asymmetric_signing"
  algorithm   = "rsa_pss_2048_sha256"
  unprotected = true
}

// Generate a signature for a message using the key created above
ephemeral "scaleway_key_manager_sign" "main" {
  key_id = scaleway_key_manager_key.main.id
  digest = "base64digest"
  region = "fr-par"
}

resource "scaleway_secret" "main" {
  name = "my-secret"
}

// Store the signature in a secret manager secret version
resource "scaleway_secret_version" "signature" {
  secret_id = scaleway_secret.main.id
  data_wo   = ephemeral.scaleway_key_manager_sign.main.signature
}

data "scaleway_secret_version" "signature" {
  secret_id = scaleway_secret.main.id
  revision  = "1"
}

// Verify the signature using the key created above, the digest and the signature retrieved from the secret
data "scaleway_key_manager_verify" "main" {
  key_id    = scaleway_key_manager_key.main.id
  region    = "fr-par"
  digest    = "base64digest"
  signature = data.scaleway_secret_version.signature.data
}
