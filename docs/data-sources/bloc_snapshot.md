---
subcategory: "Block"
page_title: "Scaleway: scaleway_block_snapshot"
---

# scaleway_block_snapshot

The `scaleway_block_snapshot` data source is used to retrieve information about a Block Storage volume snapshot.

Refer to the Block Storage [product documentation](https://www.scaleway.com/en/docs/storage/block/) and [API documentation](https://www.scaleway.com/en/developers/api/block/) for more information.

## Retrieve a volume's snapshot

The following commands allow you to:

- retrieve a snapshot specified by its name
- retrieve a snapshot specified by its name and the ID of the Block Storage volume it is associated with
- retrieve a snapshot specified by its ID

```terraform
// Get info by snapshot name
data "scaleway_block_snapshot" "my_snapshot" {
  name = "my-name"
}

// Get info by snapshot name and volume id
data "scaleway_block_snapshot" "my_snapshot" {
  name = "my-name"
  volume_id = "11111111-1111-1111-1111-111111111111"
}

// Get info by snapshot ID
data "scaleway_block_snapshot" "my_snapshot" {
  snapshot_id = "11111111-1111-1111-1111-111111111111"
}
```

## Arguments reference

This section lists the arguments that you can provide to the `scaleway_block_snapshot` data source to filter and retrieve the desired snapshot. Each argument has a specific purpose:

- `snapshot_id` - (Optional) The unique identifier of the snapshot. Only one of `name` and `snapshot_id` should be specified.
- `name` - (Optional) The name of the snapshot. Only one of name or snapshot_id should be specified.
- `volume_id` - (Optional) The unique identifier of the volume from which the snapshot was created.
- `zone` - (Defaults to the zone specified in the [provider configuration](../index.md#zone)) The [zone](../guides/regions_and_zones.md#zones) in which the snapshot exists.
- `project_id` - (Optional) The unique identifier of the Project to which the snapshot is associated.

## Attributes reference

The `scaleway_block_snapshot` data source exports certain attributes once the snapshot information is retrieved. These attributes can be referenced in other parts of your Terraform configuration. The exported attributes are those that come from the `scaleway_block_snapshot` [resource](../resources/block_snapshot.md)
