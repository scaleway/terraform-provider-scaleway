### Basic

resource "scaleway_vpc" "vpc" {
  name = "my-vpc"
}

resource "scaleway_vpc_private_network" "pn" {
  name   = "my-private-network"
  vpc_id = scaleway_vpc.vpc.id
  ipv4_subnet {
    subnet = "10.0.1.0/24"
  }
}

resource "scaleway_s2s_vpn_gateway" "gateway" {
  name               = "my-vpn-gateway"
  gateway_type       = "VGW-S"
  private_network_id = scaleway_vpc_private_network.pn.id
}
