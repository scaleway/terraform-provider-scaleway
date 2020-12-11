---
page_title: "Scaleway: scaleway_registry_image"
description: |-
  Gets information about a registry image.
---

# scaleway_registry_image

Gets information about a registry image.

## Example Usage

```hcl
# Get info by image name
data "scaleway_registry_image" "my_image" {
  name = "my-image-name"
}

# Get info by image ID
data "scaleway_registry_image" "my_image" {
  image_id = "11111111-1111-1111-1111-111111111111"
  namespace_id = "11111111-1111-1111-1111-111111111111" # Optional
}
```

## Argument Reference

- `name` - (Optional) The image name.
  Only one of `name` and `image_id` should be specified.

- `image_id` - (Optional) The image ID.
  Only one of `name` and `image_id` should be specified.

- `namespace_id` - (Optional) The namespace ID in which the image is.

- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions) in which the image exists.

- `organization_id` - (Defaults to [provider](../index.md#organization_id) `organization_id`) The ID of the organization the image is associated with.

- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the image is associated with.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the registry image.
- `size` - The size of the registry image.
- `visibility` - The privacy policy of the registry image.
- `tags` - The tags associated with the registry image
