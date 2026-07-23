### Basic

resource "scaleway_vpc_gateway_network" "main" {
  gateway_id         = scaleway_vpc_public_gateway.pg01.id
  private_network_id = scaleway_vpc_private_network.pn01.id
  dhcp_id            = scaleway_vpc_public_gateway_dhcp.dhcp01.id
  cleanup_dhcp       = true
  enable_masquerade  = true
}

data "scaleway_vpc_gateway_network" "by_id" {
  gateway_network_id = scaleway_vpc_gateway_network.main.id
}

data "scaleway_vpc_gateway_network" "by_gateway_and_pn" {
  gateway_id         = scaleway_vpc_public_gateway.pg01.id
  private_network_id = scaleway_vpc_private_network.pn01.id
}
