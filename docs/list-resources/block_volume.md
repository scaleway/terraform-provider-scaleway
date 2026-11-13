---
page_title: "Scaleway: scaleway_block_volume"
subcategory: "Block"
description: |-
  Lists Scaleway Block Storage Volumes across regions and projects.
---

# Resource: scaleway_block_volume



For more information, see [the main documentation][1].


## Example Usage

```terraform
# List block volumes filtered by name prefix
list "scaleway_block_volume" "by_name" {
  provider = scaleway

  config {
    zones = ["*"]
    name  = "my-volume"
  }
}
```

```terraform
# List block volumes filtered by tag
list "scaleway_block_volume" "by_tag" {
  provider = scaleway

  config {
    zones = ["*"]
    tags  = ["bar"]
  }
}
```

```terraform
# List block volumes in a specific zone
list "scaleway_block_volume" "by_zone" {
  provider = scaleway

  config {
    zones = ["fr-par-1"]
  }
}
```

```terraform
# List block volumes filtered by organization ID
list "scaleway_block_volume" "by_organization" {
  provider = scaleway

  config {
    zones           = ["*"]
    organization_id = "11111111-1111-1111-1111-111111111111"
  }
}
```

```terraform
# List block volumes with multiple filters combined
list "scaleway_block_volume" "combined" {
  provider = scaleway

  config {
    zones       = ["fr-par-1", "nl-ams-1"]
    project_ids = ["11111111-1111-1111-1111-111111111111"]
    tags        = ["foobar", "barfoo"]
    name        = "db-volume"
  }
}
```

```terraform
# List block volumes across all zones and all projects
list "scaleway_block_volume" "all" {
  provider = scaleway

  config {
    zones       = ["*"]
    project_ids = ["*"]
  }
}
```



## Argument Reference

The following arguments can be specified in the `config` block:

- `name` - (Optional) Name of the volume to filter for.
- `tags` - (Optional) Tags to filter for.
- `organization_id` - (Optional) Organization ID to filter for.
- `project_ids` - (Optional) Project IDs to filter for. Use `["*"]` to list
across all projects.
- `zones` - (Optional) Zones to filter for. Use `["*"]` to list from all zones.

## Attributes Reference

Each result corresponds to one Block volume and exposes the same attributes as
the [`scaleway_block_volume` resource](../resources/block_volume.md).

[1]: https://www.scaleway.com/en/docs/block-storage/concepts
