### With zone

resource "scaleway_vpc_private_network" "pn01" {
  name   = "private_network_instance"
  region = "fr-par"
}

resource "scaleway_instance_server" "base" {
  image = "ubuntu_jammy"
  type  = "DEV1-S"
  zone  = scaleway_vpc_private_network.pn01.zone
}

resource "scaleway_instance_private_nic" "pnic01" {
  server_id          = scaleway_instance_server.base.id
  private_network_id = scaleway_vpc_private_network.pn01.id
  zone               = scaleway_vpc_private_network.pn01.zone
}
