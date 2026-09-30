### Private network with static endpoint

resource "scaleway_rdb_instance" "instance" {
  name           = "rdb_instance"
  node_type      = "db-dev-s"
  engine         = "PostgreSQL-14"
  is_ha_cluster  = false
  disable_backup = true
  user_name      = "my_initial_user"
  password       = "thiZ_is_v&ry_s3cret"
}

resource "scaleway_vpc_private_network" "pn" {}

resource "scaleway_rdb_read_replica" "replica" {
  instance_id = scaleway_rdb_instance.instance.id
  private_network {
    private_network_id = scaleway_vpc_private_network.pn.id
    service_ip         = "192.168.1.254/24"
    # enable_ipam = false
  }
}
