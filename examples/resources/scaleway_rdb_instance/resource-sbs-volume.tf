### Example Block Storage Low Latency

resource "scaleway_rdb_instance" "main" {
  name              = "test-rdb-sbs"
  node_type         = "db-play2-pico"
  engine            = "PostgreSQL-15"
  is_ha_cluster     = true
  disable_backup    = true
  user_name         = "my_initial_user"
  password          = "thiZ_is_v&ry_s3cret"
  volume_type       = "sbs_15k"
  volume_size_in_gb = 10
}