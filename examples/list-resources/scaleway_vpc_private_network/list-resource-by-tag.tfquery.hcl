# List Private Networks filtered by a specific tag
list "scaleway_vpc_private_network" "by_tag" {
  provider = scaleway

  config {
    regions = ["*"]
    tags    = ["production"]
  }
}
