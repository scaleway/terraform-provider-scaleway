resource "scaleway_vpc" "vpc" {}

resource "scaleway_vpc_private_network" "pn" {
  vpc_id = scaleway_vpc.vpc.id
}

resource "scaleway_container_namespace" "with_pn" {}

resource "scaleway_container" "with_pn" {
  namespace_id       = scaleway_container_namespace.with_pn.id
  name               = "container-with-private-network"
  image              = "my-image:latest"
  private_network_id = scaleway_vpc_private_network.pn.id
}
