resource "scaleway_rdb_instance" "main" {
  name           = "test-rdb-action-read-replica-promote"
  node_type      = "db-dev-s"
  engine         = "PostgreSQL-15"
  is_ha_cluster  = false
  disable_backup = true
  user_name      = "my_initial_user"
  password       = "thiZ_is_v&ry_s3cret"
}

resource "scaleway_rdb_read_replica" "main" {
  instance_id = scaleway_rdb_instance.main.id
  direct_access {}

  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.scaleway_rdb_read_replica_promote.main]
    }
  }
}

action "scaleway_rdb_read_replica_promote" "main" {
  config {
    read_replica_id = scaleway_rdb_read_replica.main.id
    wait            = true
  }
}
