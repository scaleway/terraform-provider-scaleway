# List all buckets across all regions and all projects
list "scaleway_object_bucket" "all" {
  provider = scaleway

  config {
    regions     = ["*"]
    project_ids = ["*"]
  }
}
