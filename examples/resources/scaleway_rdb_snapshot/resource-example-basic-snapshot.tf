### Example Basic Snapshot

resource "scaleway_rdb_instance" "main" {
  name              = "test-rdb-instance"
  node_type         = "db-dev-s"
  engine            = "PostgreSQL-15"
  is_ha_cluster     = false
  disable_backup    = true
  user_name         = "my_initial_user"
  password          = "thiZ_is_v&ry_s3cret"
  tags              = ["terraform-test", "scaleway_rdb_instance", "minimal"]
  volume_type       = "sbs_5k"
  volume_size_in_gb = 10
}

resource "scaleway_rdb_snapshot" "test" {
  name        = "initial-snapshot"
  instance_id = scaleway_rdb_instance.main.id
  depends_on  = [scaleway_rdb_instance.main]
}
