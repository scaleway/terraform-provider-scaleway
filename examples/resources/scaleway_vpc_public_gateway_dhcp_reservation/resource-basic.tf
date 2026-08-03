### Basic

resource "scaleway_vpc_private_network" "main" {
  name = "your_private_network"
}

resource "scaleway_instance_server" "main" {
  image = "ubuntu_jammy"
  type  = "DEV1-S"
  zone  = "fr-par-1"

  private_network {
    pn_id = scaleway_vpc_private_network.main.id
  }
}

resource "scaleway_vpc_public_gateway_ip" "main" {
}

resource "scaleway_vpc_public_gateway_dhcp" "main" {
  subnet = "192.168.1.0/24"
}

resource "scaleway_vpc_public_gateway" "main" {
  name  = "foobar"
  type  = "VPC-GW-S"
  ip_id = scaleway_vpc_public_gateway_ip.main.id
}

resource "scaleway_vpc_gateway_network" "main" {
  gateway_id         = scaleway_vpc_public_gateway.main.id
  private_network_id = scaleway_vpc_private_network.main.id
  dhcp_id            = scaleway_vpc_public_gateway_dhcp.main.id
  cleanup_dhcp       = true
  enable_masquerade  = true
  depends_on         = [scaleway_vpc_public_gateway_ip.main, scaleway_vpc_private_network.main]
}

resource "scaleway_vpc_public_gateway_dhcp_reservation" "main" {
  gateway_network_id = scaleway_vpc_gateway_network.main.id
  mac_address        = scaleway_instance_server.main.private_network.0.mac_address
  ip_address         = "192.168.1.1"
}
