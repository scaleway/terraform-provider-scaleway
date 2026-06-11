// List secrets filtered by tags
list "scaleway_secret" "by_tags" {
  provider = scaleway

  config {
    tags = ["production", "database"]
  }
}
