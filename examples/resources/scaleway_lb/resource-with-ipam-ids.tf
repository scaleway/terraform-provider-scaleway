### With IPAM IDs

resource "scaleway_vpc" "vpc01" {
  name = "my vpc"
}

resource "scaleway_vpc_private_network" "pn01" {
  vpc_id = scaleway_vpc.vpc01.id
  ipv4_subnet {
    subnet = "172.16.32.0/22"
  }
}

resource "scaleway_ipam_ip" "ip01" {
  address = "172.16.32.7"
  source {
    private_network_id = scaleway_vpc_private_network.pn01.id
  }
}

resource "scaleway_lb_ip" "v4" {
}

resource "scaleway_lb" "lb01" {
  ip_ids = [scaleway_lb_ip.v4.id]
  name   = "my-lb"
  type   = "LB-S"

  private_network {
    private_network_id = scaleway_vpc_private_network.pn01.id
    ipam_ids           = [scaleway_ipam_ip.ip01.id]
  }
}
