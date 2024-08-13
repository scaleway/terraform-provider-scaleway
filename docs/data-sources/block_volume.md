---
subcategory: "Block"
page_title: "Scaleway: scaleway_block_volume"
---

# scaleway_block_volume

The `scaleway_block_volume` data source is used to retrieve information about a Block Storage volume.
Refer to the Block Storage [product documentation](https://www.scaleway.com/en/docs/storage/block/) and [API documentation](https://www.scaleway.com/en/developers/api/block/) for more information.

## Retrieve a Block Storage volume

The following commands allow you to:

- retrieve a Block Storage volume specified by its name
- retrieve a Block Storage volume specified by its ID


```terraform
// Get info by volume name
data "scaleway_block_volume" "my_volume" {
  name = "my-name"
}

// Get info by volume ID
data "scaleway_block_volume" "my_volume" {
  volume_id = "11111111-1111-1111-1111-111111111111"
}
```

## Argument Reference

This section lists the arguments that you can provide to the `scaleway_block_volume` data source to filter and retrieve the desired volume.

- `volume_id` - (Optional) The unique identifier of the volume. Only one of `name` and `volume_id` should be specified.
- `name` - (Optional) The name of the volume. Only one of `name` and `volume_id` should be specified.
- `zone` - (Defaults to the zone specified in the [provider configuration](../index.md#zone)). The [zone](../guides/regions_and_zones.md#zones) in which the volume exists.
- `project_id` - (Optional) The unique identifier of the Project to which the volume is associated.

## Attributes Reference

The `scaleway_block_volume` data source exports certain attributes once the volume information is retrieved. These attributes can be referenced in other parts of your Terraform configuration. The exported attributes are those that come from the `scaleway_block_volume` [resource](../resources/block_volume.md).
