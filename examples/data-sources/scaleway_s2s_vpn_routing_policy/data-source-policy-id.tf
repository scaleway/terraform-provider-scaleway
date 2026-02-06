# Get info by routing policy ID
data "scaleway_s2s_vpn_routing_policy" "my_policy" {
  routing_policy_id = "11111111-1111-1111-1111-111111111111"
}
