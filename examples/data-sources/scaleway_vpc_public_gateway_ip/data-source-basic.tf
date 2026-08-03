### Basic

resource "scaleway_vpc_public_gateway_ip" "main" {
}

data "scaleway_vpc_public_gateway_ip" "ip_by_id" {
  ip_id = scaleway_vpc_public_gateway_ip.main.id
}
