#### Default: 1 public endpoint

resource "scaleway_rdb_instance" "main" {
  node_type = "db-dev-s"
  engine    = "PostgreSQL-15"
}