// List buckets filtered by region
list "scaleway_object_bucket" "by_region" {
  provider = scaleway

  config {
    regions = ["fr-par"]
  }
}
