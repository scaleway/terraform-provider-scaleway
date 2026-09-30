### Multiple Prefixes

resource "scaleway_s2s_vpn_routing_policy" "policy" {
  name              = "my-routing-policy"
  prefix_filter_in  = ["10.0.2.0/24", "10.0.3.0/24"]
  prefix_filter_out = ["10.0.1.0/24", "172.16.0.0/16"]
}
