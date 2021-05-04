---
layout: "scaleway"
page_title: "Scaleway: scaleway_rdb_database"
description: |-
  Gets information about an RDB database.
---

# scaleway_rdb_instance

Gets information about a RDB database.

## Example Usage

```hcl
# Get the database foobar hosted on instance id 11111111-1111-1111-1111-111111111111
data "scaleway_rdb_database" "my_db" {
  instance_id = "11111111-1111-1111-1111-111111111111"
  name        = "foobar"
}
# Find the first database hosted on instance id 11111111-1111-1111-1111-111111111111
data "scaleway_rdb_database" "my_db" {
  instance_id = "11111111-1111-1111-1111-111111111111"
}
```

## Argument Reference

- `instance_id` - (Required) The RDB instance ID.

- `name` - (Optional) The name of the RDB instance.
  If omitted, the first database, in alphabethical order, will be returned


- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#zones) in which the RDB instance exists.

- `organization_id` - (Defaults to [provider](../index.md#organization_id) `organization_id`) The ID of the organization the RDB instance is in.
