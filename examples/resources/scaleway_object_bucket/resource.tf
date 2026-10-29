resource "scaleway_object_bucket" "some_bucket" {
  name = "some-unique-name"
  tags = {
    key = "value"
  }
}
