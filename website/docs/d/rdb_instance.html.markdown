---
layout: "scaleway"
page_title: "Scaleway: scaleway_rdb_instance"
description: |-
  Gets information about an RDB instance.
---

# scaleway_rdb_instance

Gets information about a RDB instance.

## Example Usage

```hcl
# Get info by IP address
data "scaleway_rdb_instance" "my_instance" {
  name = "foobar"
}

# Get info by IP ID
data "scaleway_rdb_instance" "my_instance" {
  instance_id = "11111111-1111-1111-1111-111111111111"
}
```

## Argument Reference

- `name` - (Optional) The name of the RDB instance.
  Only one of `name` and `instance_id` should be specified.

- `instance_id` - (Optional) The RDB instance ID.
  Only one of `name` and `instance_id` should be specified.

- `region` - (Defaults to [provider](../index.html#region) `region`) The [region](../guides/regions_and_zones.html#zones) in which the RDB instance exists.

- `organization_id` - (Defaults to [provider](../index.html#organization_id) `organization_id`) The ID of the organization the RDB instance is in.
