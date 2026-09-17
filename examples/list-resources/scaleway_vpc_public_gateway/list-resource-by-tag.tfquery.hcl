# List Public Gateways filtered by tag
list "scaleway_vpc_public_gateway" "by_tag" {
  provider = scaleway

  config {
    zones = ["*"]
    tags  = ["prod"]
  }
}
