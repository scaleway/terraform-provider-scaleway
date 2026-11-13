# List RDB instances filtered by tag
list "scaleway_rdb_instance" "by_tag" {
  provider = scaleway

  config {
    regions = ["*"]
    tags    = ["production"]
  }
}
