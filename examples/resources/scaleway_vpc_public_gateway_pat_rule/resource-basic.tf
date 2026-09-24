### Basic

resource "scaleway_instance_security_group" "sg01" {
  inbound_default_policy  = "drop"
  outbound_default_policy = "accept"

  inbound_rule {
    action   = "accept"
    port     = 22
    protocol = "TCP"
  }
}

resource "scaleway_instance_server" "srv01" {
  name              = "my-server"
  type              = "PLAY2-NANO"
  image             = "ubuntu_jammy"
  security_group_id = scaleway_instance_security_group.sg01.id
}

resource "scaleway_instance_private_nic" "pnic01" {
  server_id          = scaleway_instance_server.srv01.id
  private_network_id = scaleway_vpc_private_network.pn01.id
}

resource "scaleway_vpc_private_network" "pn01" {
  name = "my-pn"
}

resource "scaleway_vpc_public_gateway_dhcp" "dhcp01" {
  subnet = "192.168.0.0/24"
}

resource "scaleway_vpc_public_gateway_ip" "ip01" {}

resource "scaleway_vpc_public_gateway" "pg01" {
  name  = "my-pg"
  type  = "VPC-GW-S"
  ip_id = scaleway_vpc_public_gateway_ip.ip01.id
}

resource "scaleway_vpc_gateway_network" "gn01" {
  gateway_id         = scaleway_vpc_public_gateway.pg01.id
  private_network_id = scaleway_vpc_private_network.pn01.id
  dhcp_id            = scaleway_vpc_public_gateway_dhcp.dhcp01.id
  cleanup_dhcp       = true
  enable_masquerade  = true
}

resource "scaleway_vpc_public_gateway_dhcp_reservation" "rsv01" {
  gateway_network_id = scaleway_vpc_gateway_network.gn01.id
  mac_address        = scaleway_instance_private_nic.pnic01.mac_address
  ip_address         = "192.168.0.7"
}

# PAT rule for SSH traffic
resource "scaleway_vpc_public_gateway_pat_rule" "pat01" {
  gateway_id   = scaleway_vpc_public_gateway.pg01.id
  private_ip   = scaleway_vpc_public_gateway_dhcp_reservation.rsv01.ip_address
  private_port = 22
  public_port  = 2202
  protocol     = "tcp"
}
