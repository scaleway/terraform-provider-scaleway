# List buckets filtered by project ID
list "scaleway_object_bucket" "by_project" {
  provider = scaleway

  config {
    project_ids = ["11111111-1111-1111-1111-111111111111"]
  }
}
