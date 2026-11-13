# Get routing policy info by name
data "scaleway_interlink_routing_policy" "my_policy" {
  name = "my-routing-policy"
}
