### Simplify your rules using dynamic block and `for_each` loop

resource "scaleway_instance_security_group" "main" {
  description             = "test"
  name                    = "terraform test"
  inbound_default_policy  = "drop"
  outbound_default_policy = "accept"
}

locals {
  trusted = [
    "1.2.3.4/32",
    "4.5.6.7/32",
    "7.8.9.10/24"
  ]
}

resource "scaleway_instance_security_group_rules" "main" {
  security_group_id = scaleway_instance_security_group.main.id

  dynamic "inbound_rule" {
    for_each = local.trusted
    content {
      action   = "accept"
      ip_range = inbound_rule.value
      port     = 80
    }
  }
}
