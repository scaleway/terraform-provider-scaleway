resource "scaleway_key_manager_key" "main" {
  name        = "tf-test-kms-key-rotation-action"
  region      = "fr-par"
  usage       = "symmetric_encryption"
  algorithm   = "aes_256_gcm"
  description = "Test key"
  tags        = ["tf", "test"]
  unprotected = true

  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.scaleway_key_manager_rotate_key.main]
    }
  }
}

action "scaleway_key_manager_rotate_key" "main" {
  config {
    key_id = scaleway_key_manager_key.main.id
    region = scaleway_key_manager_key.main.region
  }
}
