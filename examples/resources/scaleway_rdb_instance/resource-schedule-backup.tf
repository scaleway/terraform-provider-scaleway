### Example with backup schedule

resource "scaleway_rdb_instance" "main" {
  name          = "test-rdb"
  node_type     = "DB-DEV-S"
  engine        = "PostgreSQL-15"
  is_ha_cluster = true
  user_name     = "my_initial_user"
  password      = "thiZ_is_v&ry_s3cret"

  disable_backup            = false
  backup_schedule_frequency = 24 # every day
  backup_schedule_retention = 7  # keep it one week
}
