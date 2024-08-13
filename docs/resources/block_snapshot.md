---
subcategory: "Block"
page_title: "Scaleway: scaleway_block_snapshot"
---

# Resource: scaleway_block_snapshot

The `scaleway_block_snapshot` resource is used to create and manage snapshots of Block Storage volumes.

Refer to the Block Storage [product documentation](https://www.scaleway.com/en/docs/storage/block/) and [API documentation](https://www.scaleway.com/en/developers/api/block/) for more information.


## Example Usage

### Create a snapshot of a Block Storage volume

The following command allows you to create a snapshot (`some-snapshot-name`) from a Block Storage volume specified by its ID.

```terraform
resource "scaleway_block_volume" "block_volume" {
  iops       = 5000
  name       = "some-volume-name"
  size_in_gb = 20
}

resource "scaleway_block_snapshot" "block_snapshot" {
  name      = "some-snapshot-name"
  volume_id = scaleway_block_volume.block_volume.id
}
```

## Arguments reference

This section lists the arguments that are supported:

- `volume_id` - (Optional) The ID of the volume to take a snapshot from.
- `name` - (Optional) The name of the snapshot. If not provided, a name will be randomly generated.
- `zone` - (Defaults to the zone specified in the [provider configuration](../index.md#zone)). The [zone](../guides/regions_and_zones.md#zones) in which the snapshot should be created.
- `project_id` - (Defaults to the Project ID specified in the [provider configuration](../index.md#project_id)). The ID of the Scaleway Project the snapshot is associated with.
- `tags` - (Optional) A list of tags to apply to the snapshot.

## Attributes reference

This section lists the attributes that are exported when the `scaleway_block_snapshot` resource is created:

- `id` - The ID of the snapshot.

~> **Important:** The IDs of Block Storage volumes snapshots are [zoned](../guides/regions_and_zones.md#resource-ids), meaning that the zone is part of the ID, in the form `{zone}/{id}`. For example, a snapshot ID migt be `fr-par-1/11111111-1111-1111-1111-111111111111`.

## Import

This section explains how to import the snapshot of a Block Storage volume using the zoned ID format (`{zone}/{id}`).

```bash
terraform import scaleway_block_snapshot.main fr-par-1/11111111-1111-1111-1111-111111111111
```
