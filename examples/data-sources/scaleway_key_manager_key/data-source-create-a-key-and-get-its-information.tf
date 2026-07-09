### Create a key and get its information

// Create a key
resource "scaleway_key_manager_key" "symmetric" {
  name        = "my-kms-key"
  region      = "fr-par"
  project_id  = "your-project-id" # optional, will use provider default if omitted
  usage       = "symmetric_encryption"
  algorithm   = "aes_256_gcm"
  description = "Key for encrypting secrets"
  tags        = ["env:prod", "kms"]
  unprotected = true

  rotation_policy {
    rotation_period = "720h" # 30 days
  }
}

// Get the key information by its ID
data "scaleway_key_manager_key" "byID" {
  key_id = "11111111-1111-1111-1111-111111111111"
}
