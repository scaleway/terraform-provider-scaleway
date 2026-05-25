# List databases filtered by name on all instances in scope
list "scaleway_rdb_database" "by_name" {
  provider = scaleway

  config {
    regions      = ["fr-par"]
    project_ids  = ["11111111-1111-1111-1111-111111111111"]
    instance_ids = ["*"]
    name         = "mydb"
  }
}
