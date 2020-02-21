---
layout: "scaleway"
page_title: "Scaleway: scaleway_registry_image_beta"
description: |-
  Gets information about a Registry Image.
---

# scaleway_registry_image_beta

Gets information about a Registry Image.

## Example Usage

```hcl
// Get info by image name
data "scaleway_registry_image_beta" "my_image" {
  name = "my-image-name"
  namespace_id = "11111111-1111-1111-1111-111111111111" // optional
}

// Get info by image ID
data "scaleway_registry_image_beta" "my_image" {
  image_id = "11111111-1111-1111-1111-111111111111"
}
```

## Argument Reference

- `name` - (Optional) The image name.
  Only one of `name` and `image_id` should be specified.

- `image_id` - (Optional) The image ID.
  Only one of `name` and `image_id` should be specified.

- `namespace_id` - (Optional) The namespace ID to filter the name on.

- `region` - (Defaults to [provider](../index.html#region) `region`) The [region](../guides/regions_and_zones.html#regions) in which the image exists.

- `organization_id` - (Defaults to [provider](../index.html#organization_id) `organization_id`) The ID of the organization the image is associated with.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the Registry Image.
- `visibility` - The Image Privacy Policy.
- `size` - The size of the Registry Image.
