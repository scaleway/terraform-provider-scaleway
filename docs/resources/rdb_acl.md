---
subcategory: "Databases"
page_title: "Scaleway: scaleway_rdb_acl"
---

# Resource: scaleway_rdb_acl

Creates and manages Scaleway Database instance authorized IPs.
For more information refer to the [API documentation](https://www.scaleway.com/en/developers/api/managed-database-postgre-mysql/#acl-rules-allowed-ips).

## Example Usage

### Basic

```terraform
resource "scaleway_rdb_instance" "main" {
  name           = "test-rdb"
  node_type      = "DB-DEV-S"
  engine         = "PostgreSQL-15"
  is_ha_cluster  = true
  disable_backup = true
  user_name      = "my_initial_user"
  password       = "thiZ_is_v&ry_s3cret"
}

resource "scaleway_rdb_acl" "main" {
  instance_id = scaleway_rdb_instance.main.id
  acl_rules {
    ip = "1.2.3.4/32"
    description = "foo"
  }
}
```

## Argument Reference

The following arguments are supported:

- `instance_id` - (Required) UUID of the Database Instance.

~> **Important:** Updates to `instance_id` will recreate the Database ACL.

- `acl_rules` - A list of ACLs (structure is described below)

- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions) in which the Database Instance should be created.

The `acl_rules` block supports:

- `ip` - (Required) The IP range to whitelist in [CIDR notation](https://en.wikipedia.org/wiki/Classless_Inter-Domain_Routing#CIDR_notation)
- `description` - (Optional) A text describing this rule. Default description: `IP allowed`

## Attributes Reference

No additional attributes are exported.

## Import

Database Instance can be imported using the `{region}/{id}`, e.g.

```bash
terraform import scaleway_rdb_acl.acl01 fr-par/11111111-1111-1111-1111-111111111111
```
