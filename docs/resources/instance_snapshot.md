---
page_title: "Scaleway: scaleway_instance_snapshot"
description: |-
Manages Scaleway Instance Snapshots.
---

# scaleway_instance_snapshot

Creates and manages Scaleway Compute Snapshots.
For more information, see [the documentation](https://developers.scaleway.com/en/products/instance/api/#snapshots-756fae).

## Example

```hcl
resource "scaleway_instance_snapshot" "main" {
    name       = "some-snapshot-name"
    volume_id  = "11111111-1111-1111-1111-111111111111"
}
```

## Arguments Reference

The following arguments are supported:

- `volume_id` - (Required) The ID of the volume to take a snapshot from.
- `name` - (Optional) The name of the snapshot. If not provided it will be randomly generated.
- `zone` - (Defaults to [provider](../index.md#zone) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which the snapshot should be created.
- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the snapshot is associated with.
- `tags` - (Optional) A list of tags to apply to the snapshot.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the snapshot.
- `size_in_gb` - (Optional) The size of the snapshot.
- `organization_id` - The organization ID the snapshot is associated with.
- `project_id` - The project ID the snapshot is associated with.
- `type` - The type of the snapshot. The possible values are: `b_ssd` (Block SSD), `l_ssd` (Local SSD).
- `created_at` - The snapshot creation time.

## Import

Snapshots can be imported using the `{zone}/{id}`, e.g.

```bash
$ terraform import scaleway_instance_snapshot.main fr-par-1/11111111-1111-1111-1111-111111111111
```
