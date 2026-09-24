### Basic

resource "scaleway_object_bucket" "some_bucket" {
  name = "unique-name"
}

resource "scaleway_object_bucket_acl" "main" {
  bucket = scaleway_object_bucket.main.id
  acl    = "private"
}
