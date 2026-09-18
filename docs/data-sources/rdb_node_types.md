---
subcategory: "Databases"
page_title: "Scaleway: scaleway_rdb_node_types"
---

# scaleway_rdb_node_types

Gets information about available RDB node types.

For further information refer to the Managed Databases for PostgreSQL and MySQL [API documentation](https://developers.scaleway.com/en/products/rdb/api/#database-instance)

## Example Usage

```terraform
data "scaleway_rdb_node_types" "all" {
  region                 = "fr-par"
  include_disabled_types = false
}

locals {
  selected_node_type = data.scaleway_rdb_node_types.all.node_types[0].name
}
```

## Argument Reference

- `region` - (Defaults to [provider](../index.md#arguments-reference) `region`) The [region](../guides/regions_and_zones.md#regions) in which to list node types.

- `include_disabled_types` - (Optional) Whether to include disabled node types in the list. Defaults to `false`.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the data source (the region).

- `node_types` - List of available RDB node types.
    - `name` - Name identifier of the node type.
    - `stock_status` - Current stock status for the node type.
    - `description` - Description of the node type offer.
    - `vcpus` - Number of virtual CPUs.
    - `memory_size_in_gb` - Amount of memory available in GB.
    - `disabled` - Whether the node type is currently disabled.
    - `beta` - Whether the node type is currently in beta.
    - `is_ha_required` - Whether the node type can only be used with high availability.
    - `generation` - Generation associated with the node type offer.
    - `instance_range` - Instance range associated with the node type offer.
    - `available_volume_types` - Available storage options for the node type.
        - `type` - Volume type.
        - `description` - Description of the volume type.
        - `min_size_in_gb` - Minimum volume size in GB.
        - `max_size_in_gb` - Maximum volume size in GB.
        - `chunk_size_in_gb` - Minimum increment level for a Block Storage volume size in GB.
        - `class` - Storage class of the volume.
