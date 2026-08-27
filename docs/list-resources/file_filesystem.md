---
page_title: "Scaleway: scaleway_file_filesystem"
subcategory: "File"
description: |-
  Lists Scaleway File File Systems across regions and projects.
---

# Resource: scaleway_file_filesystem



Lists Scaleway File File Systems.

For more information, see [the main documentation][1].


## Example Usage

```terraform
# List file systems filtered by name prefix
list "scaleway_block_volume" "by_name" {
  provider = scaleway

  config {
    regions = ["fr-par"]
    name    = "my-fs"
  }
}
```

```terraform
# List file systems filtered by tag
list "scaleway_file_filesystem" "by_tag" {
  provider = scaleway

  config {
    regions = ["fr-par"]
    tags    = ["bar"]
  }
}
```

```terraform
# List file systems in a specific region
list "scaleway_file_filesystem" "by_region" {
  provider = scaleway

  config {
    regions = ["fr-par"] # Only one supported for now
  }
}
```

```terraform
# List file systems filtered by organization ID
list "scaleway_file_filesystem" "by_organization" {
  provider = scaleway

  config {
    regions         = ["*"]
    organization_id = "11111111-1111-1111-1111-111111111111"
  }
}
```

```terraform
# List file systems with multiple filters combined
list "scaleway_file_filesystem" "combined" {
  provider = scaleway

  config {
    regions     = ["fr-par"]
    project_ids = ["11111111-1111-1111-1111-111111111111"]
    tags        = ["foobar", "barfoo"]
    name        = "db-volume"
  }
}
```



## Argument Reference

The following arguments can be specified in the `config` block:

- `name` - (Optional) Name of the snapshot to filter on.
- `tags` - (Optional) Tags to filter on.
- `organization_id` - (Optional) Organization ID to filter on.
- `project_ids` - (Optional) Project IDs to filter on. Use `["*"]` to list
across all projects.
- `filesystem_ids` - (Optional) Filesystem IDs to filter on.
- `regions` - (Optional) Regions to filter for. for now, only `fr-par` is
supported.

~> **Important:** Filesystems are only available on the `fr-par` region. You
must specify `["fr-par"]` in the `regions` field for the list to work.

## Attributes Reference

Each result corresponds to one file system and exposes the same attributes as
the [`scaleway_file_filesystem` resource](../resources/file_filesystem.md).

[1]: https://www.scaleway.com/en/docs/file-storage/concepts/
