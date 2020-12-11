---
page_title: "Scaleway: scaleway_instance_volume"
description: |-
  Manages Scaleway Compute Instance Volumes.
---

# scaleway_instance_volume

Creates and manages Scaleway Compute Instance Volumes. For more information, see [the documentation](https://developers.scaleway.com/en/products/instance/api/#volumes-7e8a39).

## Example

```hcl
resource "scaleway_instance_volume" "server_volume" {
    type       = "l_ssd"
    name       = "some-volume-name"
    size_in_gb = 20
}
```

## Arguments Reference

The following arguments are supported:

- `type` - (Required) The type of the volume. The possible values are: `b_ssd` (Block SSD), `l_ssd` (Local SSD).
- `size_in_gb` - (Optional) The size of the volume. Only one of `size_in_gb`, `from_volume_id` and `from_volume_id` should be specified.
- `from_volume_id` - (Optional) If set, the new volume will be copied from this volume. Only one of `size_in_gb`, `from_volume_id` and `from_snapshot_id` should be specified.
- ``from_snapshot_id`` - (Optional) If set, the new volume will be created from this snapshot. Only one of `size_in_gb`, `from_volume_id` and `from_snapshot_id` should be specified.
- `name` - (Optional) The name of the volume. If not provided it will be randomly generated.
- `zone` - (Defaults to [provider](../index.md#zone) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which the volume should be created.
- `organization_id` - (Defaults to [provider](../index.md#organization_id) `organization_id`) The ID of the organization the volume is associated with.
- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the volume is associated with.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the volume.
- `server_id` - The id of the associated server.

## Import

volumes can be imported using the `{zone}/{id}`, e.g.

```bash
$ terraform import scaleway_instance_volume.server_volume fr-par-1/11111111-1111-1111-1111-111111111111
```
