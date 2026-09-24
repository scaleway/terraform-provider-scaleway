## Retrieve an Object Storage bucket

resource "scaleway_object_bucket" "main" {
  name = "bucket.test.com"
  tags = {
    foo = "bar"
  }
}

data "scaleway_object_bucket" "selected" {
  name = scaleway_object_bucket.main.id
}
