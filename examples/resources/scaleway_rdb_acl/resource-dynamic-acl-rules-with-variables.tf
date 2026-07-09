### Dynamic ACL Rules with Variables

variable "allowed_ips" {
  description = "Map of allowed IPs with descriptions"
  type        = map(string)
  default = {
    "1.2.3.4/32"  = "Office IP"
    "5.6.7.8/32"  = "Home IP"
    "10.0.0.0/24" = "Internal network"
  }
}

resource "scaleway_rdb_acl" "main" {
  instance_id = scaleway_rdb_instance.main.id

  dynamic "acl_rules" {
    for_each = var.allowed_ips
    content {
      ip          = acl_rules.key
      description = acl_rules.value
    }
  }
}
