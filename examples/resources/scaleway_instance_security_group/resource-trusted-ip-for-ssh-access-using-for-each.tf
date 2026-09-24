### Trusted IP for SSH access (using for_each)

locals {
  trusted = ["192.168.0.1", "192.168.0.2", "192.168.0.3"]
}

resource "scaleway_instance_security_group" "dummy" {
  inbound_default_policy  = "drop"
  outbound_default_policy = "accept"

  dynamic "inbound_rule" {
    for_each = local.trusted

    content {
      action   = "accept"
      port     = 22
      ip_range = inbound_rule.value
    }
  }
}
