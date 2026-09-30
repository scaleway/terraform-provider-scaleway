### Basic

resource "scaleway_object_bucket" "test" {
  name = "my-bucket"
  acl  = "public-read"
}

resource "scaleway_object" "some_file" {
  bucket       = scaleway_object_bucket.test.name
  key          = "index.html"
  file         = "index.html"
  visibility   = "public-read"
  content_type = "text/html"
}

resource "scaleway_object_bucket_website_configuration" "test" {
  bucket = scaleway_object_bucket.test.name
  index_document {
    suffix = "index.html"
  }
  error_document {
    key = "error.html"
  }
}
