resource "scaleway_rdb_instance" "main" {
  name              = "test-rdb-action-snapshot"
  node_type         = "db-dev-s"
  engine            = "PostgreSQL-15"
  is_ha_cluster     = false
  disable_backup    = true
  user_name         = "my_initial_user"
  password          = "thiZ_is_v&ry_s3cret"
  volume_type       = "sbs_5k"
  volume_size_in_gb = 10

  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.scaleway_rdb_instance_snapshot.main]
    }
  }
}

action "scaleway_rdb_instance_snapshot" "main" {
  config {
    instance_id = scaleway_rdb_instance.main.id
    name        = "tf-rdb-snapshot"
    wait        = true
  }
}
