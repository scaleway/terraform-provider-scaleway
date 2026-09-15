### With subnets

resource "scaleway_vpc_private_network" "pn_priv" {
  name = "subnet_demo"
  tags = ["demo", "terraform"]

  ipv4_subnet {
    subnet = "192.168.0.0/24"
  }
  ipv6_subnets {
    subnet = "fd46:78ab:30b8:177c::/64"
  }
  ipv6_subnets {
    subnet = "fd46:78ab:30b8:c7df::/64"
  }
}
