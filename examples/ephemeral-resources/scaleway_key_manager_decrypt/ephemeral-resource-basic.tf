resource "scaleway_key_manager_key" "main" {
  name        = "tf-test-decrypt-key"
  region      = "fr-par"
  usage       = "symmetric_encryption"
  algorithm   = "aes_256_gcm"
  unprotected = true
}

ephemeral "scaleway_key_manager_decrypt" "test_decrypt" {
  key_id     = scaleway_key_manager_key.main.id
  ciphertext = ephemeral.scaleway_key_manager_encrypt.test_encrypt.ciphertext
  region     = "fr-par"
  depends_on = [ephemeral.scaleway_key_manager_encrypt.test_encrypt]
}
