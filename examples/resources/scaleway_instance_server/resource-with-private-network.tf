### With private network

resource "scaleway_vpc_private_network" "pn01" {
  name = "private_network_instance"
}

resource "scaleway_instance_server" "base" {
  image = "ubuntu_jammy"
  type  = "DEV1-S"

  private_network {
    pn_id = scaleway_vpc_private_network.pn01.id
  }
}
