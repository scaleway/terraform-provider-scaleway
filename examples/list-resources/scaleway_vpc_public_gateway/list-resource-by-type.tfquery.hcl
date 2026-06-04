# List Public Gateways filtered by gateway type
list "scaleway_vpc_public_gateway" "by_type" {
  provider = scaleway

  config {
    zones = ["*"]
    types = ["VPC-GW-S"]
  }
}
