# List all Private Networks belonging to a specific VPC
list "scaleway_vpc_private_network" "by_vpc" {
  provider = scaleway

  config {
    regions = ["fr-par"]
    vpc_id  = "11111111-1111-1111-1111-111111111111"
  }
}
