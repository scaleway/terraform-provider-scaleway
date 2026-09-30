### Multiple ACL Rules

resource "scaleway_rdb_acl" "main" {
  instance_id = scaleway_rdb_instance.main.id

  acl_rules {
    ip          = "1.2.3.4/32"
    description = "Office IP"
  }

  acl_rules {
    ip          = "5.6.7.8/32"
    description = "Home IP"
  }

  acl_rules {
    ip          = "10.0.0.0/24"
    description = "Internal network"
  }
}
