### Basic

resource "scaleway_vpc_public_gateway_dhcp" "main" {
  subnet = "192.168.1.0/24"
}
