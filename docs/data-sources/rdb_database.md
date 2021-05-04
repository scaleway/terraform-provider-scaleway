---
layout: "scaleway"
page_title: "Scaleway: scaleway_rdb_database"
description: |-
  Gets information about an RDB database.
---

# scaleway_rdb_database

Gets information about a RDB database.

## Example Usage

```hcl
# Get the database foobar hosted on instance id 11111111-1111-1111-1111-111111111111
data "scaleway_rdb_database" "my_db" {
  instance_id = "11111111-1111-1111-1111-111111111111"
  name        = "foobar"
}
```

## Argument Reference

- `instance_id` - (Required) The RDB instance ID.

- `name` - (Required) The name of the RDB instance.

- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#zones) in which the RDB instance exists.

- `organization_id` - (Defaults to [provider](../index.md#organization_id) `organization_id`) The ID of the organization the RDB instance is in.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `owner` - The name of the owner of the database.
- `managed` - Whether or not the database is managed or not.
- `size` - Size of the database (in bytes).
