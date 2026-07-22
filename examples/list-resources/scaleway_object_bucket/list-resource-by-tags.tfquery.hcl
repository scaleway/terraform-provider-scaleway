# List buckets filtered by tags
list "scaleway_object_bucket" "by_tags" {
  provider = scaleway

  config {
    tags = ["production", "env:prod"]
  }
}
