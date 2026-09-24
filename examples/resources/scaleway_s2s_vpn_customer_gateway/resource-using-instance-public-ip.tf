### Using Instance Public IP

resource "scaleway_instance_ip" "vpn_endpoint_ip" {}

resource "scaleway_instance_server" "vpn_endpoint" {
  name   = "vpn-endpoint"
  type   = "DEV1-S"
  image  = "ubuntu_jammy"
  ip_ids = [scaleway_instance_ip.vpn_endpoint_ip.id]
}

resource "scaleway_s2s_vpn_customer_gateway" "customer_gw" {
  name        = "my-customer-gateway"
  ipv4_public = scaleway_instance_ip.vpn_endpoint_ip.address
  asn         = 65000
}
