### Basic

resource "scaleway_sdb_sql_database" "database" {
  name    = "my-database"
  min_cpu = 0
  max_cpu = 8
}
