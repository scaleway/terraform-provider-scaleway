resource "scaleway_interlink_routing_policy" "main" {
  name              = "my-routing-policy"
  prefix_filter_in  = ["10.0.2.0/24"]
  prefix_filter_out = ["10.0.1.0/24"]
}
