resource "scaleway_vpc" "vpc01" {
  name   = "my-vpc"
  region = "nl-ams"
}

resource "scaleway_vpc_private_network" "pn01" {
  name   = "my-private-network"
  vpc_id = scaleway_vpc.vpc01.id
  region = "nl-ams"
}

resource "scaleway_vpc_ingress_rule" "main" {
  vpc_id                     = scaleway_vpc.vpc01.id
  source                     = "10.0.0.0/24"
  nexthop_private_network_id = scaleway_vpc_private_network.pn01.id
  nexthop_resource_ip        = "10.0.0.10"
  region                     = "nl-ams"
}
