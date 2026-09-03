## Example Static and PAT Rule

resource "scaleway_vpc_private_network" "main" {}

resource "scaleway_instance_security_group" "main" {
  inbound_default_policy  = "drop"
  outbound_default_policy = "accept"

  inbound_rule {
    action = "accept"
    port   = "22"
  }
}

resource "scaleway_instance_server" "main" {
  image = "ubuntu_jammy"
  type  = "DEV1-S"
  zone  = "fr-par-1"

  security_group_id = scaleway_instance_security_group.main.id
}

resource "scaleway_instance_private_nic" "main" {
  server_id          = scaleway_instance_server.main.id
  private_network_id = scaleway_vpc_private_network.main.id
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
}

resource "scaleway_vpc_public_gateway_dhcp_reservation" "main" {
  gateway_network_id = scaleway_vpc_gateway_network.main.id
  mac_address        = scaleway_instance_private_nic.main.mac_address
  ip_address         = "192.168.1.4"
}

### VPC PAT RULE
resource "scaleway_vpc_public_gateway_pat_rule" "main" {
  gateway_id   = scaleway_vpc_public_gateway.main.id
  private_ip   = scaleway_vpc_public_gateway_dhcp_reservation.main.ip_address
  private_port = 22
  public_port  = 2222
  protocol     = "tcp"
}

data "scaleway_vpc_public_gateway_dhcp_reservation" "by_id" {
  reservation_id = scaleway_vpc_public_gateway_dhcp_reservation.main.id
}
