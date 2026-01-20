### Examples of endpoint configuration

##### Database Instances can have a maximum of 1 public endpoint and 1 private endpoint. They can have both, or none.

#### 1 static Private Network endpoint

resource "scaleway_vpc_private_network" "pn" {
  ipv4_subnet {
    subnet = "172.16.20.0/22"
  }
}

resource "scaleway_rdb_instance" "main" {
  node_type = "db-dev-s"
  engine    = "PostgreSQL-15"
  private_network {
    pn_id  = scaleway_vpc_private_network.pn.id
    ip_net = "172.16.20.4/22" # IP address within a given IP network
    # enable_ipam = false
  }
}
