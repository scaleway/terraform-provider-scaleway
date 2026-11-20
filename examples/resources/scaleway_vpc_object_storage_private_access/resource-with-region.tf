resource "scaleway_vpc" "vpc01" {
  name   = "my-vpc"
  region = "nl-ams"
}

resource "scaleway_vpc_private_network" "pn01" {
  name   = "my-private-network"
  vpc_id = scaleway_vpc.vpc01.id
  region = "nl-ams"
}

resource "scaleway_vpc_object_storage_private_access" "main" {
  vpc_id = scaleway_vpc.vpc01.id
  region = "nl-ams"

  private_network_ids = [
    scaleway_vpc_private_network.pn01.id,
  ]
}
