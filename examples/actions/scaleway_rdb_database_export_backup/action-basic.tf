resource "scaleway_rdb_instance" "main" {
  name              = "test-rdb-action-backup-export"
  node_type         = "db-dev-s"
  engine            = "PostgreSQL-15"
  is_ha_cluster     = false
  disable_backup    = true
  user_name         = "my_initial_user"
  password          = "thiZ_is_v&ry_s3cret"
  volume_type       = "sbs_5k"
  volume_size_in_gb = 10
}

resource "scaleway_rdb_database" "main" {
  instance_id = scaleway_rdb_instance.main.id
  name        = "test_db"
}

resource "scaleway_rdb_database_backup" "main" {
  instance_id   = scaleway_rdb_instance.main.id
  database_name = scaleway_rdb_database.main.name
  name          = "test-backup-export"
  depends_on    = [scaleway_rdb_database.main]

  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.scaleway_rdb_database_export_backup.main]
    }
  }
}

action "scaleway_rdb_database_export_backup" "main" {
  config {
    backup_id = scaleway_rdb_database_backup.main.id
    wait      = true
  }
}
