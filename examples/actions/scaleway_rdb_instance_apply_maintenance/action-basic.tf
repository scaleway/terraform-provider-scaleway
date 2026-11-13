resource "scaleway_rdb_instance" "main" {
  name           = "my-rdb-instance"
  node_type      = "db-dev-s"
  engine         = "PostgreSQL-15"
  is_ha_cluster  = false
  disable_backup = true
  user_name      = "my_initial_user"
  password       = "thiZ_is_v&ry_s3cret"

  lifecycle {
    action_trigger {
      events  = [after_update]
      actions = [action.scaleway_rdb_instance_apply_maintenance.main]
    }
  }
}

action "scaleway_rdb_instance_apply_maintenance" "main" {
  config {
    instance_id = scaleway_rdb_instance.main.id
    wait        = true
  }
}
