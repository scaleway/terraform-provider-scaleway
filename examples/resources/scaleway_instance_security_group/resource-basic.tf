### Basic

resource "scaleway_instance_security_group" "allow_all" {
}

resource "scaleway_instance_security_group" "web" {
  inbound_default_policy = "drop" # By default we drop incoming traffic that do not match any inbound_rule

  inbound_rule {
    action   = "accept"
    port     = 22
    ip_range = "212.47.225.64/32"
  }

  inbound_rule {
    action = "accept"
    port   = 80
  }

  inbound_rule {
    action     = "accept"
    protocol   = "UDP"
    port_range = "22-23"
  }
}
