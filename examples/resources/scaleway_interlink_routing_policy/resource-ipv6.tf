resource "scaleway_interlink_routing_policy" "main" {
  name              = "my-routing-policy-v6"
  is_ipv6           = true
  prefix_filter_in  = ["2001:db8:1::/48"]
  prefix_filter_out = ["2001:db8:2::/48"]
}
