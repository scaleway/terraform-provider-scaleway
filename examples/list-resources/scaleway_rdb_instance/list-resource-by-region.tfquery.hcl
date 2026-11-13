# List RDB instances in a specific region for a specific project
list "scaleway_rdb_instance" "region" {
  provider = scaleway

  config {
    regions     = ["fr-par"]
    project_ids = ["11111111-1111-1111-1111-111111111111"]
  }
}
