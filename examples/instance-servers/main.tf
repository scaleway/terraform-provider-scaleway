provider "scaleway" {}

resource "scaleway_vpc" "vpc" {}

resource "scaleway_vpc_private_network" "pn" {
  vpc_id = scaleway_vpc.vpc.id
}

resource "scaleway_instance_ip" "ip" {
  count = var.server_count
}

resource "scaleway_instance_server" "server" {
  count = var.server_count
  type = "PLAY2-MICRO"
  image = "ubuntu_jammy"
  ip_ids = [scaleway_instance_ip.ip[count.index].id]
}
