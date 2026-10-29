# List RDB instances filtered by name prefix
list "scaleway_rdb_instance" "by_name" {
  provider = scaleway

  config {
    regions = ["*"]
    name    = "my-rdb"
  }
}
