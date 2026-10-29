resource "scaleway_object_bucket" "test" {
  name   = "my-unique-bucket-name"
  region = "fr-par"
}

resource "scaleway_object_bucket_server_side_encryption_configuration" "test" {
  bucket = scaleway_object_bucket.test.name
  region = "fr-par"

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}
