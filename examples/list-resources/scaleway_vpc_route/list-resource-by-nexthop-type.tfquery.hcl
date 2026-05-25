# List VPC routes filtered by nexthop resource type
list "scaleway_vpc_route" "by_nexthop_type" {
  provider = scaleway

  config {
    regions              = ["fr-par"]
    nexthop_resource_type = "instance_private_nic"
  }
}
