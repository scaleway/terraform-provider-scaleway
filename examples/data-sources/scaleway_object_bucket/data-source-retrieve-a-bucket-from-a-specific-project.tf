## Retrieve a bucket from a specific project

data "scaleway_object_bucket" "selected" {
  name       = "bucket.test.com"
  project_id = "11111111-1111-1111-1111-111111111111"
}
