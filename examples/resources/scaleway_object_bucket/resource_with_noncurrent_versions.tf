resource "scaleway_object_bucket" "main" {
  name   = "mybuckectid"
  region = "fr-par"

  lifecycle_rule {
    id      = "id1"
    prefix  = "path1/"
    enabled = true

    noncurrent_version_expiration {
      noncurrent_days = 90
    }

    noncurrent_version_transition {
      noncurrent_days = 30
      storage_class   = "ONEZONE_IA"
    }

    noncurrent_version_transition {
      noncurrent_days = 60
      storage_class   = "GLACIER"
    }
  }
}
