# List RDB instances across all regions and all projects
list "scaleway_rdb_instance" "all" {
  provider = scaleway

  config {
    regions     = ["*"]
    project_ids = ["*"]
  }
}
