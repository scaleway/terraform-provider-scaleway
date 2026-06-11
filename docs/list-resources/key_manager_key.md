---
page_title: "Scaleway: scaleway_key_manager_key"
subcategory: "Key Manager"
description: |-
  Lists Scaleway Key Manager Keys across regions and projects.
---

# Resource: scaleway_key_manager_key

Lists Scaleway Key Manager Keys across regions and projects.

For more information, see [the main documentation](https://www.scaleway.com/en/docs/key-manager/concepts/).

## Example Usage

```terraform
# List all keys across all regions and projects
list "scaleway_keymanager_key" "all" {
  provider = scaleway

  config {
    regions     = ["*"]
    project_ids = ["*"]
  }
}
```
```terraform
# List keys filtered by name
list "scaleway_keymanager_key" "by_name" {
  provider = scaleway

  config {
    name = "my-key"
  }
}
```
```terraform
# List keys in specific projects
list "scaleway_keymanager_key" "by_projects" {
  provider = scaleway

  config {
    project_ids = [
      "11111111-1111-1111-1111-111111111111",
      "22222222-2222-2222-2222-222222222222",
    ]
  }
}
```

## Argument Reference

The following arguments can be specified in the `config` block:

- `regions` - (Optional) Regions to filter for. Use `["*"]` to list from all regions. If not specified, the provider default region is used.
- `project_ids` - (Optional) Project IDs to filter for. Use `["*"]` to list across all projects. If not specified, the provider default project is used.
- `tags` - (Optional) Tags to filter for.
- `name` - (Optional) Name of the key to filter for.
- `usage` - (Optional) Usage of the key to filter for. Possible values: `symmetric_encryption`, `asymmetric_encryption`, `asymmetric_signing`.
- `scheduled_for_deletion` - (Optional) Filter keys by deletion status.

## Attributes Reference

In addition to the arguments above, the following attributes are exported for each Key:

- `id` - The ID of the key.
- `name` - The name of the key.
- `project_id` - The project ID the key belongs to.
- `region` - The region where the key is stored.
- `usage` - The key usage type.
- `algorithm` - The algorithm used for the key.
- `description` - The description of the key.
- `tags` - The list of the key's tags.
- `rotation_count` - The rotation count tracks the number of times the key has been rotated.
- `created_at` - The date and time of the creation of the key.
- `updated_at` - The date and time of the last update of the key.
- `protected` - Returns true if key protection is applied to the key.
- `locked` - Returns true if the key is locked.
- `rotated_at` - The date and time of the last rotation of the key.
- `rotation_policy` - The key rotation policy.
- `state` - The state of the key.
