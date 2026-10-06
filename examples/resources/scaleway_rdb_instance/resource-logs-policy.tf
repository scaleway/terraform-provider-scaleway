### Example with logs policy

resource "scaleway_rdb_instance" "main" {
  name          = "test-rdb"
  node_type     = "DB-DEV-S"
  engine        = "PostgreSQL-15"
  is_ha_cluster = true
  user_name     = "my_initial_user"
  password      = "thiZ_is_v&ry_s3cret"

  logs_policy {
    max_age_retention    = 30
    total_disk_retention = 100000000 # in bytes
  }
}
