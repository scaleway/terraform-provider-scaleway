resource "scaleway_rdb_instance" "main" {
  name           = "rdb-instance-logs-details"
  node_type      = "db-dev-s"
  engine         = "PostgreSQL-15"
  is_ha_cluster  = false
  disable_backup = true
  user_name      = "my_initial_user"
  password       = "thiZ_is_v&ry_s3cret"
}

data "scaleway_rdb_instance_logs_details" "details" {
  instance_id = scaleway_rdb_instance.main.id
}
