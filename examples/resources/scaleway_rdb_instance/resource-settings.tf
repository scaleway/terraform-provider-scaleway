### Example with Settings

resource "scaleway_rdb_instance" "main" {
  name           = "test-rdb"
  node_type      = "db-dev-s"
  disable_backup = true
  engine         = "MySQL-8"
  user_name      = "my_initial_user"
  password       = "thiZ_is_v&ry_s3cret"
  init_settings = {
    "lower_case_table_names" = 1
  }
  settings = {
    "max_connections" = "350"
  }
}
