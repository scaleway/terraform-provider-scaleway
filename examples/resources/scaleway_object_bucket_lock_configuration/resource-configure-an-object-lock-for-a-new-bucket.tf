### Configure an Object Lock for a new bucket

resource "scaleway_object_bucket" "main" {
  name = "MyBucket"
  acl  = "public-read"

  object_lock_enabled = true
}

resource "scaleway_object_bucket_lock_configuration" "main" {
  bucket = scaleway_object_bucket.main.name

  rule {
    default_retention {
      mode = "GOVERNANCE"
      days = 1
    }
  }
}
