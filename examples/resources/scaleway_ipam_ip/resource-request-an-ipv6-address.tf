### Request an IPv6 address

resource "scaleway_vpc" "vpc01" {
  name = "my vpc"
}

resource "scaleway_vpc_private_network" "pn01" {
  vpc_id = scaleway_vpc.vpc01.id
  ipv6_subnets {
    subnet = "fd46:78ab:30b8:177c::/64"
  }
}

resource "scaleway_ipam_ip" "ip01" {
  is_ipv6 = true
  source {
    private_network_id = scaleway_vpc_private_network.pn01.id
  }
}
