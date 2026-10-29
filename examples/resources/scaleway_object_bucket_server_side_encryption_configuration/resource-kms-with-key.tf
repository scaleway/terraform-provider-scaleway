resource "scaleway_object_bucket" "test" {
  name   = "my-bucket"
  region = "fr-par"
}

resource "scaleway_key_manager_key" "mykey" {
  name        = "my-kms-key"
  description = "This key is used to encrypt bucket objects"
  usage       = "asymmetric_encryption"
  algorithm   = "rsa_oaep_4096_sha256"
  unprotected = "true"
}

resource "scaleway_object_bucket_server_side_encryption_configuration" "test" {
  bucket = scaleway_object_bucket.test.name
  region = "fr-par"

  rule {
    apply_server_side_encryption_by_default {
      kms_master_key_id = scaleway_key_manager_key.mykey.name
      sse_algorithm     = "aws:kms"
    }
    bucket_key_enabled = true
  }
}

