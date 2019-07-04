---
layout: "scaleway"
page_title: "Scaleway: scaleway_compute_instance_volume"
sidebar_current: "docs-scaleway-resource-compute-instance-volume"
description: |-
  Manages Scaleway Compute Instance Volumes.
---

# scaleway_compute_instance_volume

Creates and manages Scaleway Compute Instance Volumes. For more information, see [the documentation](https://developers.scaleway.com/en/products/instance/api/#volumes-7e8a39).

## Example Usage

```hcl
resource "scaleway_compute_instance_volume" "server_volume" {
    type       = "l_ssd"
	name       = "some-volume-name"
	size_in_gb = 20
}
```

## Arguments Reference

The following arguments are supported:

- `type` - (Optional) Type type of the volume: `b_ssd` (Block SSD), `l_ssd` (Local SSD) or `l_hdd` (Local HDD). Defaults to `b_ssd`.
- `size_in_gb` - (Optional) The size of the volume (leave this empty when using `from_volume_id` or `from_snapshot_id`).
- `from_volume_id` - (Optional) If set, the new volume will be copied from this volume. (leave this empty when using `size_in_gb` or `from_snapshot_id`).
- `from_snapshot_id` - (Optional) If set, the new volume will be created from this snapshot. (leave this empty when using `size_in_gb` or `from_volume_id`).
- `name` - (Optional) The name of the volume. If not provided it will be randomly generated.
- `zone` - (Optional) The [zone](https://developers.scaleway.com/en/quickstart/#zone-definition) in which the volume should be created. If it is not provided, the `zone` of the provider is used.
- `project_id` - (Optional) The ID of the project the volume is associated with. If it is not provided, the provider `project_id` is used.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the volume.
- `name` - The name of the volume.
- `type`- The type of the volume: Block SSD, Local SSD or Local HDD.
- `size_in_gb` - The size of the volume.
- `from_volume_id` - The base volume this volume is copied from.
- `from_snapshot_id` - The snapshot this volume is created from.
- `server_id` - The id of the associated server.
- `zone` - The [zone](https://developers.scaleway.com/en/quickstart/#zone-definition) of the volume.
- `project_id` - The ID of the project the volume is associated with.

## Import

volumes can be imported using the `{zone}/{id}`, e.g.

```bash
$ terraform import scaleway_compute_instance_volume.server_volume fr-par-1/11111111-1111-1111-1111-111111111111
```
