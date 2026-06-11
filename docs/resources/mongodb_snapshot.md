---
subcategory: "MongoDB®"
page_title: "Scaleway: scaleway_mongodb_snapshot"
---

# Resource: scaleway_mongodb_snapshot

Creates and manages Scaleway MongoDB® snapshots.
For more information refer to the [product documentation](https://www.scaleway.com/en/docs/managed-mongodb-databases/).

## Example Usage

```terraform

resource "scaleway_mongodb_snapshot" "main" {
  instance_id = "${scaleway_mongodb_instance.main.id}"
  name        = "name-snapshot"
  expires_at  = "2024-12-31T23:59:59Z"
}
```

## Argument Reference

The following arguments are supported:

- `instance_id` - (Required) The ID of the MongoDB® instance from which the snapshot was created.

- `name` - (Optional) The name of the MongoDB® snapshot.

- `expires_at` - (Required) The expiration date of the MongoDB® snapshot in ISO 8601 format (e.g. `2024-12-31T23:59:59Z`).

~> **Important:** Once set, `expires_at` cannot be removed.

- `region` - (Defaults to [provider](../index.md#arguments-reference) `region`) The [region](../guides/regions_and_zones.md#regions) in which the MongoDB® snapshot should be created.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the snapshot.

- `instance_name` - The name of the MongoDB® instance from which the snapshot was created.

- `size` - The size of the MongoDB® snapshot in bytes.

- `node_type` - The type of node associated with the MongoDB® snapshot.

- `volume_type` - The type of volume used for the MongoDB® snapshot.

- `created_at` - The date and time when the MongoDB® snapshot was created.

- `updated_at` - The date and time of the last update of the MongoDB® snapshot.

## Import

MongoDB® snapshots can be imported using the `{region}/{id}`, e.g.

```bash
terraform import scaleway_mongodb_snapshot.main fr-par/11111111-1111-1111-1111-111111111111
```
