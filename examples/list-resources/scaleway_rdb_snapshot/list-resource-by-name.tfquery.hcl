# List snapshots filtered by name
list "scaleway_rdb_snapshot" "by_name" {
  provider = scaleway

  config {
    regions      = ["fr-par"]
    project_ids  = ["11111111-1111-1111-1111-111111111111"]
    instance_ids = ["*"]
    name         = "my-snapshot"
  }
}
