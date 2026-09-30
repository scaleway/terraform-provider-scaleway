### Basic

resource "scaleway_s2s_vpn_customer_gateway" "customer_gw" {
  name        = "my-customer-gateway"
  ipv4_public = "203.0.113.1"
  asn         = 65000
}
