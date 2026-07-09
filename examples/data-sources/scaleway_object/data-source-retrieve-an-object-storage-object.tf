## Retrieve an Object Storage object

resource "scaleway_object_bucket" "main" {
  name = "bucket.test.com"
}

resource "scaleway_object" "example" {
  bucket  = scaleway_object_bucket.main.name
  key     = "example.txt"
  content = "Hello world!"
}

data "scaleway_object" "selected" {
  bucket = scaleway_object.example.bucket
  key    = scaleway_object.example.key
}
