### Basic

resource "scaleway_vpc_public_gateway_dhcp" "main" {
  subnet = "192.168.0.0/24"
}

data "scaleway_vpc_public_gateway_dhcp" "dhcp_by_id" {
  dhcp_id = scaleway_vpc_public_gateway_dhcp.main.id
}
