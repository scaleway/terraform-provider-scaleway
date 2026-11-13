data "scaleway_sdb_sql_versions" "pg" {
  name = "16"
}

locals {
  selected_version = data.scaleway_sdb_sql_versions.pg.versions[0].name
}
