---
layout: "scaleway"
page_title: "Scaleway: scaleway_rdb_acl"
description: |-
  Gets information about the RDB instance network Access Control List.
---

# scaleway_rdb_acl

Gets information about the RDB instance network Access Control List.

## Example Usage

```hcl
# Get the database ACL for the instanceid 11111111-1111-1111-1111-111111111111
data "scaleway_rdb_acl" "my_acl" {
  instance_id = "11111111-1111-1111-1111-111111111111"
}
```

## Argument Reference

- `instance_id` - (Required) The RDB instance ID.

## Attribute Reference

In addition to all above arguments, the following attributes are exported:

- `acl_rules` - A list of ACLs (structure is described below)

The `acl_rules` block supports:

- `ip` - The ip range to whitelist in [CIDR notation](https://en.wikipedia.org/wiki/Classless_Inter-Domain_Routing#CIDR_notation)
- `description` - A simple text describing this rule
