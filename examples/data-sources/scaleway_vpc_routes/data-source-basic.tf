### Basic

resource "scaleway_vpc" "vpc01" {
  name           = "tf-vpc-route"
  enable_routing = true
}

resource "scaleway_vpc_private_network" "pn01" {
  name   = "tf-pn-route"
  vpc_id = scaleway_vpc.vpc01.id
}

resource "scaleway_vpc_private_network" "pn02" {
  name   = "tf-pn_route-2"
  vpc_id = scaleway_vpc.vpc01.id
}

# Find routes with a matching VPC ID
data "scaleway_vpc_routes" "routes_by_vpc_id" {
  vpc_id = scaleway_vpc.vpc01.id
}

# Find routes with a matching next hop private network ID
data "scaleway_vpc_routes" "routes_by_pn_id" {
  vpc_id                     = scaleway_vpc.vpc01.id
  nexthop_private_network_id = scaleway_vpc_private_network.pn01.id
}

# Find routes with an IPv6 destination 
data "scaleway_vpc_routes" "routes_by_pn_id" {
  vpc_id  = scaleway_vpc.vpc01.id
  is_ipv6 = true
}

# Find routes with a nexthop resource type
data "scaleway_vpc_routes" "routes_by_pn_id" {
  vpc_id                = scaleway_vpc.vpc01.id
  nexthop_resource_type = "vpc_gateway_network"
}
