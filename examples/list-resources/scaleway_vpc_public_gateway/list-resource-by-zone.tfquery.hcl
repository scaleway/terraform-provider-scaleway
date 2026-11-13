# List Public Gateways in a specific zone
list "scaleway_vpc_public_gateway" "by_zone" {
  provider = scaleway

  config {
    zones = ["fr-par-1"]
  }
}
