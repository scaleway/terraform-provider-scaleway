---
page_title: "Scaleway: scaleway_rdb_instance"
subcategory: "Databases"
description: |-
  Lists Scaleway RDB instances across regions and projects.
---

# Resource: scaleway_rdb_instance



For more information, see the [product documentation](https://www.scaleway.com/en/docs/managed-databases/).


## Example Usage

```terraform
# List RDB instances across all regions and all projects
list "scaleway_rdb_instance" "all" {
  provider = scaleway

  config {
    regions     = ["*"]
    project_ids = ["*"]
  }
}
```

```terraform
# List RDB instances filtered by name prefix
list "scaleway_rdb_instance" "by_name" {
  provider = scaleway

  config {
    regions = ["*"]
    name    = "my-rdb"
  }
}
```

```terraform
# List RDB instances in a specific region for a specific project
list "scaleway_rdb_instance" "region" {
  provider = scaleway

  config {
    regions     = ["fr-par"]
    project_ids = ["11111111-1111-1111-1111-111111111111"]
  }
}
```

```terraform
# List RDB instances filtered by tag
list "scaleway_rdb_instance" "by_tag" {
  provider = scaleway

  config {
    regions = ["*"]
    tags    = ["production"]
  }
}
```

```terraform
# List RDB instances with scheduled maintenances
list "scaleway_rdb_instance" "with_maintenance" {
  provider = scaleway

  config {
    regions          = ["*"]
    has_maintenances = true
  }
}
```



## Argument Reference

The following arguments can be specified in the `config` block:

- `name` - (Optional) Name of the RDB instance to filter on.
- `tags` - (Optional) Tags of the RDB instance to filter on.
- `organization_id` - (Optional) Organization ID of the RDB instance to filter on.
- `project_ids` - (Optional) Project IDs to filter for. Use `["*"]` to list across all projects.
- `regions` - (Optional) Regions to filter for. Use `["*"]` to list from all regions.
- `has_maintenances` - (Optional) Whether to only list instances with scheduled maintenances.

## Attributes Reference

Each result corresponds to one RDB instance and exposes the same attributes as the [`scaleway_rdb_instance` resource](../resources/rdb_instance.md).
