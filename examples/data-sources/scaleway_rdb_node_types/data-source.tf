data "scaleway_rdb_node_types" "all" {
  region                 = "fr-par"
  include_disabled_types = false
}

locals {
  selected_node_type = data.scaleway_rdb_node_types.all.node_types[0].name
}
