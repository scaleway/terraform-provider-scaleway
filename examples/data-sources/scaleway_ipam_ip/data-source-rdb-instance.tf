### RDB instance

# Find the private IPv4 using resource name
resource "scaleway_vpc_private_network" "pn" {}

resource "scaleway_rdb_instance" "main" {
  name           = "test-rdb"
  node_type      = "DB-DEV-S"
  engine         = "PostgreSQL-15"
  is_ha_cluster  = true
  disable_backup = true
  user_name      = "my_initial_user"
  password       = "thiZ_is_v&ry_s3cret"
  private_network {
    pn_id = scaleway_vpc_private_network.pn.id
  }
}

data "scaleway_ipam_ip" "by_name" {
  resource {
    name = scaleway_rdb_instance.main.name
    type = "rdb_instance"
  }
  type = "ipv4"
}
