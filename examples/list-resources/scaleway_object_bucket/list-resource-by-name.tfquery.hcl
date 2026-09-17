# List buckets across all regions, filtered by name prefix
list "scaleway_object_bucket" "by_name" {
  provider = scaleway

  config {
    regions = ["*"]
    name    = "my-bucket"
  }
}
