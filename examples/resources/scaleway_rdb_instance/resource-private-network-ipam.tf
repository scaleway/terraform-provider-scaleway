#### 1 IPAM Private Network endpoint + 1 public endpoint

resource "scaleway_vpc_private_network" "pn" {}

resource "scaleway_rdb_instance" "main" {
  node_type = "DB-DEV-S"
  engine    = "PostgreSQL-15"
  private_network {
    pn_id       = scaleway_vpc_private_network.pn.id
    enable_ipam = true
  }
  load_balancer {}
}