---
page_title: "Scaleway: scaleway_block_snapshot"
subcategory: "Block"
description: |-
  Lists Scaleway Block Storage Snapshots across regions and projects.
---

# Resource: scaleway_block_snapshot



For more information, see [the main documentation][1].

## Argument Reference

The following arguments can be specified in the `config` block:

- `name` - (Optional) Name of the snapshot to filter for.
- `tags` - (Optional) Tags to filter for.
- `volume_id` - (Optional) Volume IDs to filter for. Use `["*"]` only to include
all volumes in each selected zone and project. Otherwise each value must be a
zonal ID (`zone/uuid`) or a bare volume UUID.
- `organization_id` - (Optional) Organization ID to filter for.
- `project_ids` - (Optional) Project IDs to filter for. Use `["*"]` to list
across all projects.
- `zones` - (Optional) Zones to filter for. Use `["*"]` to list from all zones.
- `include_deleted` - (Optional) Display deleted snapshots not erased yet.

## Attributes Reference

Each result corresponds to one Block snapshot and exposes the same attributes as
the [`scaleway_block_snapshot` resource](../resources/block_snapshot.md).

[1]: https://www.scaleway.com/en/docs/block-storage/concepts
