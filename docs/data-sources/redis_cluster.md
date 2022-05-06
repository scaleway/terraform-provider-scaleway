---
layout: "scaleway"
page_title: "Scaleway: scaleway_redis_instance"
description: |-
Gets information about a Redis cluster.
---

# scaleway_redis_cluster

Gets information about a Redis cluster.

## Example Usage

```hcl
# Get info by name
data "scaleway_redis_cluster" "my_cluster" {
  name = "foobar"
}

# Get info by cluster ID
data "scaleway_redis_cluster" "my_cluster" {
  cluster_id = "11111111-1111-1111-1111-111111111111"
}
```

## Argument Reference

- `name` - (Optional) The name of the Redis cluster.
  Only one of the `name` and `cluster_id` should be specified.

- `cluster_id` - (Optional) The Redis cluster ID.
  Only one of the `name` and `cluster_id` should be specified.

- `zone` - (Default to [provider](../index.md) `region`) The [zone](../guides/regions_and_zones.md#zones) in which the server exists.

- `project_id` - (Default to [provider](../index.md) `project_id`)
