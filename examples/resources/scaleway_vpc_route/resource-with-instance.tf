### With Instance

resource "scaleway_vpc" "vpc01" {
  name = "tf-vpc-vpn"
}

resource "scaleway_vpc_private_network" "pn01" {
  name = "tf-pn-vpn"
  ipv4_subnet {
    subnet = "172.16.64.0/22"
  }
  vpc_id = scaleway_vpc.vpc01.id
}

resource "scaleway_instance_server" "server01" {
  name  = "tf-server-vpn"
  type  = "PLAY2-MICRO"
  image = "openvpn"
}

resource "scaleway_instance_private_nic" "pnic01" {
  private_network_id = scaleway_vpc_private_network.pn01.id
  server_id          = scaleway_instance_server.server01.id
}

resource "scaleway_vpc_route" "rt01" {
  vpc_id              = scaleway_vpc.vpc01.id
  description         = "tf-route-vpn"
  tags                = ["tf", "route"]
  destination         = "10.0.0.0/24"
  nexthop_resource_id = scaleway_instance_private_nic.pnic01.id
}
