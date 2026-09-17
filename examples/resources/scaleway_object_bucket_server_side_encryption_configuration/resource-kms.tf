resource "scaleway_object_bucket" "test" {
  name   = "my-bucket"
  region = "fr-par"
}

resource "scaleway_object_bucket_server_side_encryption_configuration" "test" {
  bucket = scaleway_object_bucket.test.name
  region = "fr-par"

  rule {
    apply_server_side_encryption_by_default {
      kms_master_key_id = "my-key-id"
      sse_algorithm     = "aws:kms"
    }
    bucket_key_enabled = true
  }
}

