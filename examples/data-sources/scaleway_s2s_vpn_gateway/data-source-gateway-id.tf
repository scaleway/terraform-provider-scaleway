# Get info by gateway ID
data "scaleway_s2s_vpn_gateway" "my_gateway" {
  vpn_gateway_id = "11111111-1111-1111-1111-111111111111"
}
