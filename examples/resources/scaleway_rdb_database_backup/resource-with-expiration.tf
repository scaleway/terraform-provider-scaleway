### With expiration

resource "scaleway_rdb_database_backup" "main" {
  instance_id   = data.scaleway_rdb_instance.main.id
  database_name = data.scaleway_rdb_database.main.name
  expires_at    = "2022-06-16T07:48:44Z"
}
