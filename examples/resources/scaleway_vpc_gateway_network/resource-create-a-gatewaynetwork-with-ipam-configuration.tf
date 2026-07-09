### Create a GatewayNetwork with IPAM configuration

resource "scaleway_vpc" "vpc01" {
  name = "my vpc"
}

resource "scaleway_vpc_private_network" "pn01" {
  name = "pn_test_network"
  ipv4_subnet {
    subnet = "172.16.64.0/22"
  }
  vpc_id = scaleway_vpc.vpc01.id
}

resource "scaleway_vpc_public_gateway" "pg01" {
  name = "foobar"
  type = "VPC-GW-S"
}

resource "scaleway_vpc_gateway_network" "main" {
  gateway_id         = scaleway_vpc_public_gateway.pg01.id
  private_network_id = scaleway_vpc_private_network.pn01.id
  enable_masquerade  = true
  ipam_config {
    push_default_route = true
  }
}
