---
subcategory: "Container Registry"
page_title: "Scaleway: scaleway_registry_image_tag"
---

# scaleway_registry_image_tag

Gets information about a specific tag of a Container Registry image.

## Example Usage

```hcl
# Get info by tag ID and image ID
data "scaleway_registry_image_tag" "my_image_tag" {
    tag_id  = "11111111-1111-1111-1111-111111111111"
    image_id = "22222222-2222-2222-2222-222222222222"
}

```

## Argument Reference

- `tag_id` - (Required) The ID of the registry image tag.

- `image_id` - (Required) The ID of the registry image.

- `region` - (Defaults to provider region) The region in which the registry image tag exists.

- `organization_id` - (Defaults to provider organization_id) The ID of the organization the image tag is associated with.

- `project_id` - (Defaults to provider project_id) The ID of the project the image tag is associated with.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the registry image.

~> **Important:** Registry images' IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111`

- `name` - The name of the registry image tag.

- `status` - The status of the registry image tag.

- `digest` - Hash of the tag content. Several tags of the same image may have the same digest.

- `created_at` - The date and time when the registry image tag was created.

- `updated_at` - The date and time of the last update to the registry image tag.

- `endpoint` - The endpoint where the registry image tag is accessible.

- `organization_id` - The organization ID the image tag is associated with.