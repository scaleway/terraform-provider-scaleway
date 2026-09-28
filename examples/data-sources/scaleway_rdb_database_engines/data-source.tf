data "scaleway_rdb_database_engines" "pg" {
  region  = "fr-par"
  name    = "PostgreSQL"
  version = "16"
}

locals {
  selected_engine = data.scaleway_rdb_database_engines.pg.engines[0].versions[0].name
}
