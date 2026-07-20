---
page_title: "Scaleway: scaleway_block_volume"
subcategory: "Block"
description: |-
  Lists Scaleway Block Storage Volumes across regions and projects.
---

# Resource: scaleway_block_volume



For more information, see [the main documentation][1].

## Argument Reference

The following arguments can be specified in the `config` block:

- `name` - (Optional) Name of the volume to filter for.
- `tags` - (Optional) Tags to filter for.
- `organization_id` - (Optional) Organization ID to filter for.
- `project_ids` - (Optional) Project IDs to filter for. Use `["*"]` to list
across all projects.
- `zones` - (Optional) Zones to filter for. Use `["*"]` to list from all zones.

## Attributes Reference

Each result corresponds to one Block volume and exposes the same attributes as
the [`scaleway_block_volume` resource](../resources/block_volume.md).

[1]: https://www.scaleway.com/en/docs/block-storage/concepts
