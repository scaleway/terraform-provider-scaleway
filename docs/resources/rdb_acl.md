---
page_title: "Scaleway: scaleway_rdb_acl"
description: |-
  Manages Scaleway Database ACL rules.
---

# scaleway_rdb_acl

Creates and manages Scaleway Database instance authorized IPs.
For more information, see [the documentation](https://developers.scaleway.com/en/products/rdb/api/#acl-rules-allowed-ips).

## Examples

### Basic

```hcl
resource "scaleway_rdb_acl" "main" {
  instance_id = scaleway_rdb_instance.main.id
  acl_rules {
    ip = "1.2.3.4/32"
    description = "foo"
  }
}
```

## Arguments Reference

The following arguments are supported:

- `instance_id` - (Required) The instance on which to create the ACL.

~> **Important:** Updates to `instance_id` will recreate the Database ACL.

- `acl_rules` - A list of ACLs (structure is described below)

The `acl_rules` block supports:

- `ip` - (Required) The ip range to whitelist in [CIDR notation](https://en.wikipedia.org/wiki/Classless_Inter-Domain_Routing#CIDR_notation)
- `description` - (Optional) A simple text describing this rule. Default description: `IP allowed`


## Attributes Reference

All arguments above are exported.

## Import

Database Instance can be imported using the `{region}/{id}`, e.g.

```bash
$ terraform import scaleway_rdb_acl.acl01 fr-par/11111111-1111-1111-1111-111111111111
```

