### Basic

resource "scaleway_rdb_instance" "instance" {
  name           = "test-rdb-rr-update"
  node_type      = "db-dev-s"
  engine         = "PostgreSQL-14"
  is_ha_cluster  = false
  disable_backup = true
  user_name      = "my_initial_user"
  password       = "thiZ_is_v&ry_s3cret"
  tags           = ["terraform-test", "scaleway_rdb_read_replica", "minimal"]
}

resource "scaleway_rdb_read_replica" "replica" {
  instance_id = scaleway_rdb_instance.instance.id
  direct_access {}
}
