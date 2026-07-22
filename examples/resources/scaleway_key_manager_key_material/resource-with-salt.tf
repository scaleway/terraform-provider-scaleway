resource "scaleway_key_manager_key" "main" {
  name        = "my-external-key"
  description = "Key with externally imported material and salt"
  usage       = "symmetric_encryption"
  algorithm   = "aes_256_gcm"
  origin      = "external"
  region      = "fr-par"
}

resource "random_bytes" "key_material" {
  length = 32 # 256-bit key for AES-256
}

resource "random_bytes" "salt" {
  length = 16 # 128-bit salt
}

resource "scaleway_key_manager_key_material" "main" {
  key_id                  = scaleway_key_manager_key.main.id
  key_material_wo         = base64encode(random_bytes.key_material.base64)
  key_material_wo_version = 1
  salt_wo                 = base64encode(random_bytes.salt.base64)
  salt_wo_version         = 1
}
