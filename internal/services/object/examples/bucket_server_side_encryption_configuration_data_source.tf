# Get bucket server side encryption configuration by ID

data "scaleway_object_bucket_server_side_encryption_configuration" "by_id" {
  bucket_server_side_encryption_configuration_id = scaleway_object_bucket_server_side_encryption_configuration.main.id
}

# Get bucket server side encryption configuration by bucket name

data "scaleway_object_bucket_server_side_encryption_configuration" "by_bucket" {
  bucket = scaleway_object_bucket.main.name
}

# Example usage with a bucket and configuration

resource "scaleway_object_bucket" "main" {
  name = "my-bucket"
  region = "fr-par"
}

resource "scaleway_object_bucket_server_side_encryption_configuration" "main" {
  bucket = scaleway_object_bucket.main.name

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}
