# List database backups on a specific RDB instance
list "scaleway_rdb_database_backup" "by_instance" {
  provider = scaleway

  config {
    regions      = ["fr-par"]
    project_ids  = ["11111111-1111-1111-1111-111111111111"]
    instance_ids = ["fr-par/22222222-2222-2222-2222-222222222222"]
  }
}
