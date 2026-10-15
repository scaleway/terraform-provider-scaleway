### Example Clone

resource "scaleway_rdb_instance" "source" {
  name           = "test-rdb-source"
  node_type      = "DB-DEV-S"
  engine         = "PostgreSQL-15"
  is_ha_cluster  = false
  disable_backup = true
  user_name      = "my_initial_user"
  password       = "thiZ_is_v&ry_s3cret"
}

resource "scaleway_rdb_instance" "clone" {
  name       = "test-rdb-clone"
  node_type  = "DB-DEV-M"
  clone_from = scaleway_rdb_instance.source.id

  # Optional: set explicitly to override what CloneInstance inherits from the source
  # is_ha_cluster  = false
  # disable_backup = true
  # volume_type    = "lssd"

  # Keep the public endpoint provisioned by CloneInstance
  load_balancer {}
}
