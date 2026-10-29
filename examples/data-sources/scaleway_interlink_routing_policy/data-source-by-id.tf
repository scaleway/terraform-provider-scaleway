# Get routing policy info by ID
data "scaleway_interlink_routing_policy" "my_policy" {
  routing_policy_id = "11111111-1111-1111-1111-111111111111"
}
