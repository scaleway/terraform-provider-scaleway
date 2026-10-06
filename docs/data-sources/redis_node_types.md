---
subcategory: "Redis"
page_title: "Scaleway: scaleway_redis_node_types"
---

# scaleway_redis_node_types

Gets information about available Redis™ node types.

For further information refer to the Managed Database for Redis™ [API documentation](https://developers.scaleway.com/en/products/redis/api/v1alpha1/#clusters-a85816).

## Example Usage

```terraform
data "scaleway_redis_node_types" "available" {
  zone                   = "fr-par-1"
  include_disabled_types = false
}

locals {
  selected_node_type = data.scaleway_redis_node_types.available.node_types[0].name
}
```

## Argument Reference

- `zone` - (Optional, Computed, Defaults to [provider](../index.md#arguments-reference) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which to list node types.

- `include_disabled_types` - (Optional) Whether to include disabled node types in the list. Defaults to `false`.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the data source (the zone).

- `node_types` - List of available Redis™ node types.
    - `name` - Name of the node type.
    - `stock_status` - Current stock status of the node type.
    - `description` - Current specifications of the offer.
    - `vcpus` - Number of virtual CPUs.
    - `memory_size_in_gb` - Amount of memory available in GB.
    - `disabled` - Whether the node type is currently disabled.
    - `beta` - Whether the node type is currently in beta.
