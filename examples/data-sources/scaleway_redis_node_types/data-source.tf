data "scaleway_redis_node_types" "available" {
  zone                   = "fr-par-1"
  include_disabled_types = false
}

locals {
  selected_node_type = data.scaleway_redis_node_types.available.node_types[0].name
}
