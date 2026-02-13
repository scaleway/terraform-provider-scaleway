### Basic object creation

resource "scaleway_object_bucket" "some_bucket" {
  name = "some-unique-name"
}

resource "scaleway_object" "some_file" {
  bucket = scaleway_object_bucket.some_bucket.id
  key    = "object_path"

  file = "myfile"
  hash = filemd5("myfile")
}
