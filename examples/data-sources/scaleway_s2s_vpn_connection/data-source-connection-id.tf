# Get info by connection ID
data "scaleway_s2s_vpn_connection" "my_connection" {
  connection_id = "11111111-1111-1111-1111-111111111111"
}
