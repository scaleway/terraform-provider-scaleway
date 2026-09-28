---
subcategory: "Block"
page_title: "Scaleway: scaleway_block_snapshots"
---

# scaleway_block_snapshots

Gets information about multiple Block Storage volume snapshots.

For more information, see the [main documentation](https://www.scaleway.com/en/docs/block-storage/) or [API documentation](https://www.scaleway.com/en/developers/api/block/).

## Example Usage

```terraform
# List all snapshots in a zone
data "scaleway_block_snapshots" "all" {
  zone = "fr-par-1"
}

# Filter snapshots by name
data "scaleway_block_snapshots" "by_name" {
  name = "my-snapshot"
  zone = "fr-par-1"
}

# Filter snapshots by tags
data "scaleway_block_snapshots" "by_tags" {
  tags = ["backup", "daily"]
  zone = "fr-par-1"
}

# Filter snapshots by the volume they were created from
data "scaleway_block_snapshots" "by_volume" {
  volume_id = scaleway_block_volume.main.id
}
```

## Argument Reference

- `name` - (Optional) Filter snapshots by their name. Snapshots with a matching name are listed.

- `tags` - (Optional) List of tags used as filter. Snapshots with one or more matching tags are listed.

- `volume_id` - (Optional) The ID of the volume the snapshots were created from, used as filter. Can be a bare UUID or a zoned ID (`zone/uuid`).

- `order_by` - (Optional) Criteria to use when ordering the list of snapshots. Possible values are: `created_at_asc` (default), `created_at_desc`, `name_asc` and `name_desc`. Use `created_at_desc` to get the most recent snapshot first.

- `zone` - (Optional, Computed, Defaults to [provider](../index.md#arguments-reference) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which snapshots exist.

- `project_id` - (Optional) The ID of the Project the snapshots are associated with, used as filter.

- `organization_id` - (Optional) The ID of the Organization the snapshots are associated with, used as filter.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `snapshots` - List of snapshots matching the filters. Empty when no snapshot matches.
    - `id` - The ID of the snapshot in the `zone/uuid` format.
    - `name` - The name of the snapshot.
    - `volume_id` - The ID of the volume from which the snapshot was created. Empty if the parent volume was deleted.
    - `tags` - The list of tags assigned to the snapshot.
    - `srn` - The Scaleway Resource Name (SRN) of the snapshot.
    - `zone` - The [zone](../guides/regions_and_zones.md#zones) in which the snapshot is.
    - `project_id` - The ID of the Project the snapshot is associated with.
