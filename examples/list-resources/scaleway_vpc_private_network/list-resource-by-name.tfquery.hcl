# List Private Networks across all regions filtered by name prefix
list "scaleway_vpc_private_network" "by_name" {
  provider = scaleway

  config {
    regions     = ["*"]
    name        = "my-network"
  }
}
