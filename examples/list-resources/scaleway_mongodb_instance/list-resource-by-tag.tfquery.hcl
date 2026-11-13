# List MongoDB instances filtered by tag
list "scaleway_mongodb_instance" "by_tag" {
  provider = scaleway

  config {
    regions = ["*"]
    tags    = ["production"]
  }
}
