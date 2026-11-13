---
page_title: "Scaleway: scaleway_mongodb_instance"
subcategory: "MongoDB®"
description: |-
  Lists Scaleway MongoDB® instances across regions and projects.
---

# Resource: scaleway_mongodb_instance



For more information, see the [product documentation](https://www.scaleway.com/en/docs/managed-mongodb-databases/).


## Example Usage

```terraform
# List MongoDB instances across all regions and all projects
list "scaleway_mongodb_instance" "all" {
  provider = scaleway

  config {
    regions     = ["*"]
    project_ids = ["*"]
  }
}
```

```terraform
# List MongoDB instances filtered by name prefix
list "scaleway_mongodb_instance" "by_name" {
  provider = scaleway

  config {
    regions = ["*"]
    name    = "my-mongodb"
  }
}
```

```terraform
# List MongoDB instances in a specific region for a specific project
list "scaleway_mongodb_instance" "region" {
  provider = scaleway

  config {
    regions     = ["fr-par"]
    project_ids = ["11111111-1111-1111-1111-111111111111"]
  }
}
```

```terraform
# List MongoDB instances filtered by tag
list "scaleway_mongodb_instance" "by_tag" {
  provider = scaleway

  config {
    regions = ["*"]
    tags    = ["production"]
  }
}
```



## Argument Reference

The following arguments can be specified in the `config` block:

- `name` - (Optional) Name of the MongoDB instance to filter on.
- `tags` - (Optional) Tags of the MongoDB instance to filter on.
- `organization_id` - (Optional) Organization ID of the MongoDB instance to filter on.
- `project_ids` - (Optional) Project IDs to filter for. Use `["*"]` to list across all projects.
- `regions` - (Optional) Regions to filter for. Use `["*"]` to list from all regions.

## Attributes Reference

Each result corresponds to one MongoDB instance and exposes the same attributes as the [`scaleway_mongodb_instance` resource](../resources/mongodb_instance.md).
