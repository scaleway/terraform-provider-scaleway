# Retrieve a VPC ingress rule by filters
data "scaleway_vpc_ingress_rule" "by_pn" {
  nexthop_private_network_id = "fr-par/11111111-1111-1111-1111-111111111111"
}
