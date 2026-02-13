### Using Write-Only SSE Customer Key

resource "scaleway_object_bucket" "encrypted_bucket" {
  name = "encrypted-bucket"
}

# Generate an ephemeral encryption key (not stored in the state)
ephemeral "random_password" "encryption_key" {
  length      = 32
  special     = false
  upper       = true
  lower       = true
  numeric     = true
  min_upper   = 1
  min_lower   = 1
  min_numeric = 1
  # Only hex characters for SSE-C keys
  override_special = ""
}

resource "scaleway_object" "encrypted_file" {
  bucket  = scaleway_object_bucket.encrypted_bucket.id
  key     = "secret-file"
  content = "This is a secret content"

  # Use write-only encryption key
  sse_customer_key_wo         = ephemeral.random_password.encryption_key.result
  sse_customer_key_wo_version = 1
}
