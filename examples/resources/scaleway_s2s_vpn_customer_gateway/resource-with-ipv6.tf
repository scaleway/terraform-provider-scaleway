### With IPv6

resource "scaleway_s2s_vpn_customer_gateway" "customer_gw" {
  name        = "my-customer-gateway"
  ipv4_public = "203.0.113.1"
  ipv6_public = "2001:db8::1"
  asn         = 65000
}
