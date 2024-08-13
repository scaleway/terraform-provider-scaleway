---
subcategory: "Block"
page_title: "Scaleway: scaleway_block_volume"
---

# Resource: scaleway_block_volume

The `scaleway_block_volume` resource is used to create and manage Scaleway Block Storage volumes.

Refer to the Block Storage [product documentation](https://www.scaleway.com/en/docs/storage/block/) and [API documentation](https://www.scaleway.com/en/developers/api/block/) for more information.


## Example Usage

### Create a Block Storage volume

The following command allows you to create a Block Storage volume of 20 GB with a 5000 [IOPS](https://www.scaleway.com/en/docs/storage/block/concepts/#iops).

```terraform
resource "scaleway_block_volume" "block_volume" {
  iops       = 5000
  name       = "some-volume-name"
  size_in_gb = 20
}
```

### With snapshot

```terraform
resource "scaleway_block_volume" "base" {
  name       = "block-volume-base"
  iops       = 5000
  size_in_gb = 20
}

resource "scaleway_block_snapshot" "main" {
  name      = "block-volume-from-snapshot"
  volume_id = scaleway_block_volume.base.id
}

resource "scaleway_block_volume" "main" {
  name        = "block-volume-from-snapshot"
  iops        = 5000
  snapshot_id = scaleway_block_snapshot.main.id
}
```

## Arguments reference

This section lists the arguments that are supported:

- `iops` - (Required) The maximum [IOPs](https://www.scaleway.com/en/docs/storage/block/concepts/#iops) expected, must match available options.
- `name` - (Optional) The name of the volume. If not provided, a name will be randomly generated.
- `size_in_gb` - (Optional) The size of the volume in gigabytes. Only one of `size_in_gb`, and `snapshot_id` should be specified.
- `snapshot_id` - (Optional) If set, the new volume will be created from this snapshot. Only one of `size_in_gb`, `snapshot_id` should be specified.
- `tags` - (Optional) A list of tags to apply to the volume.
- `zone` - (Defaults to the zone specified in the [provider configuration](../index.md#zone)). The [zone](../guides/regions_and_zones.md#zones) in which the volume should be created.
- `project_id` - (Defaults to the Project ID specified in the [provider configurqtion](../index.md#project_id)). The ID of the Project the volume is associated with.

## Attributes reference

This section lists the attributes that are exported when the `scaleway_block_volume` resource is created:

- `id` - The ID of the volume.

~> **Important:** The IDs of Block Storage volumes are [zoned](../guides/regions_and_zones.md#resource-ids), meaning that the zone is part of the ID, in the `{zone}/{id}` format. For example, a volume ID might look like the following: `fr-par-1/11111111-1111-1111-1111-111111111111`.

- `organization_id` - The Organization ID the volume is associated with.

## Import

This section explains how to import a Block Storage volume using the zoned ID (`{zone}/{id}`) format.

```bash
terraform import scaleway_block_volume.block_volume fr-par-1/11111111-1111-1111-1111-111111111111
```
