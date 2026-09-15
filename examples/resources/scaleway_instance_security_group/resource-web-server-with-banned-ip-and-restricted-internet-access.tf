### Web server with banned IP and restricted internet access

resource "scaleway_instance_security_group" "web" {
  inbound_default_policy  = "drop" # By default we drop incoming traffic that do not match any inbound_rule.
  outbound_default_policy = "drop" # By default we drop outgoing traffic that do not match any outbound_rule.

  inbound_rule {
    action   = "drop"
    ip_range = "1.1.1.1/32" # Banned IP range
  }

  inbound_rule {
    action   = "accept"
    port     = 22
    ip_range = "212.47.225.64/32"
  }

  inbound_rule {
    action = "accept"
    port   = 443
  }

  outbound_rule {
    action   = "accept"
    ip_range = "8.8.8.8/32" # Only allow outgoing connection to this IP range.
  }
}
