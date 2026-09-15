### With IPAM IP IDs

resource "scaleway_vpc" "vpc01" {
  name = "vpc_instance"
}

resource "scaleway_vpc_private_network" "pn01" {
  name = "private_network_instance"
  ipv4_subnet {
    subnet = "172.16.64.0/22"
  }
  vpc_id = scaleway_vpc.vpc01.id
}

resource "scaleway_ipam_ip" "ip01" {
  address = "172.16.64.7"
  source {
    private_network_id = scaleway_vpc_private_network.pn01.id
  }
}

resource "scaleway_instance_server" "server01" {
  image = "ubuntu_focal"
  type  = "PLAY2-MICRO"
}

resource "scaleway_instance_private_nic" "pnic01" {
  private_network_id = scaleway_vpc_private_network.pn01.id
  server_id          = scaleway_instance_server.server01.id
  ipam_ip_ids        = [scaleway_ipam_ip.ip01.id]
}
