resource "scaleway_object_bucket" "test" {
  name = "my-unique-bucket-name"
}

resource "scaleway_object_bucket_server_side_encryption_configuration" "test" {
  bucket = scaleway_object_bucket.test.name

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}
