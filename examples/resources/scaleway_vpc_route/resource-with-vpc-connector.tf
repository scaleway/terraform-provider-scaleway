### With VPC Connector

resource "scaleway_vpc" "vpc01" {
  name = "tf-vpc-source"
}

resource "scaleway_vpc" "vpc02" {
  name = "tf-vpc-target"
}

resource "scaleway_vpc_connector" "main" {
  name          = "tf-conn-route"
  vpc_id        = scaleway_vpc.vpc01.id
  target_vpc_id = scaleway_vpc.vpc02.id
}

resource "scaleway_vpc_route" "rt01" {
  vpc_id                   = scaleway_vpc.vpc01.id
  description              = "tf-route-connector"
  tags                     = ["tf", "route"]
  destination              = "10.0.0.0/24"
  nexthop_vpc_connector_id = scaleway_vpc_connector.main.id
}
