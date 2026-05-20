# List databases on all RDB instances in a region and project
list "scaleway_rdb_database" "all_instances" {
  provider = scaleway

  config {
    regions      = ["fr-par"]
    project_ids  = ["11111111-1111-1111-1111-111111111111"]
    instance_ids = ["*"]
  }
}
