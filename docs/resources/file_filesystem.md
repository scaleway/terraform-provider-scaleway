---
subcategory: "File"
page_title: "Scaleway: scaleway_file_filesystem"
---

# Resource: scaleway_file_filesystem

-> **This product is currently in private beta. To request access, please contact your Technical Account Manager.**

Creates and manages a Scaleway File Storage filesystem in a specific region. A filesystem is a scalable storage resource that can be mounted on Compute instances and is typically used for share persistent storage between multiple instances (RWX).

This resource allows you to define and manage the size, tags, and region of a filesystem, and track its creation and update timestamps, current status, and number of active attachments.

## Example Usage

### Basic

```terraform
resource scaleway_file_filesystem file {
  name = "my-nfs-filesystem"
  size_in_gb = 100
}
```

## Argument Reference

- `name` - (Optional) The name of the filesystem. If not provided, a random name will be generated.
- `size_in_gb` - (Required) The size of the filesystem in bytes, with a granularity of 100 GB (10¹¹ bytes).
      - Minimum: 100 GB (100000000000 bytes)
      - Maximum: 10 TB (10000000000000 bytes)
- `tags` - (Optional) A list of tags associated with the filesystem.
- `region` - (Defaults to [provider](../index.md#region) `region`) The region where the filesystem will be created (e.g., fr-par, nl-ams).
- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the server is
  associated with.
- `organization_id` - (Defaults to [provider](../index.md#organization_id) `organization_id`) The ID of the organization the user is associated with.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the filesystem.
- `status` - The current status of the filesystem. Possible values include creating, available, etc.
- `number_of_attachments` - The number of active attachments (mounts) on the filesystem.
- `created_at` - The date and time when the File Storage filesystem was created.
- `updated_at` - The date and time of the last update to the File Storage filesystem.

## Import

File Storage filesystems can be imported using the `{region}/{id}`, e.g.

```bash
terraform import scaleway_file_filesystem.main fr-par/11111111-1111-1111-1111-111111111111
```
