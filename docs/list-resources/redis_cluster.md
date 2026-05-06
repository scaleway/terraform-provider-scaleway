---
page_title: "Scaleway: scaleway_redis_cluster"
subcategory: "Redis"
description: |-
  Lists Scaleway Redis clusters across zones and projects.
---

# Resource: scaleway_redis_cluster



For more information, see the [product documentation](https://www.scaleway.com/en/docs/managed-databases/redis/).


## Example Usage

```terraform
# List Redis clusters across all zones and all projects
list "scaleway_redis_cluster" "all" {
  provider = scaleway

  config {
    zones       = ["*"]
    project_ids = ["*"]
  }
}
```

```terraform
# List Redis clusters filtered by name prefix
list "scaleway_redis_cluster" "by_name" {
  provider = scaleway

  config {
    zones = ["*"]
    name  = "my-redis"
  }
}
```

```terraform
# List Redis clusters in a specific zone for a specific project
list "scaleway_redis_cluster" "zone" {
  provider = scaleway

  config {
    zones       = ["fr-par-2"]
    project_ids = ["11111111-1111-1111-1111-111111111111"]
  }
}
```

```terraform
# List Redis clusters filtered by tag
list "scaleway_redis_cluster" "by_tag" {
  provider = scaleway

  config {
    zones = ["*"]
    tags  = ["production"]
  }
}
```

```terraform
# List Redis clusters filtered by engine version
list "scaleway_redis_cluster" "by_version" {
  provider = scaleway

  config {
    zones       = ["*"]
    project_ids = ["*"]
    version     = "7.2"
  }
}
```



## Argument Reference

The following arguments can be specified in the `config` block:

- `name` - (Optional) Name of the Redis cluster to filter on.
- `tags` - (Optional) Tags of the Redis cluster to filter on.
- `organization_id` - (Optional) Organization ID of the Redis cluster to filter on.
- `project_ids` - (Optional) Project IDs to filter for. Use `["*"]` to list across all projects.
- `zones` - (Optional) Zones to filter for. Use `["*"]` to list from all zones.
- `version` - (Optional) Redis engine version to filter on.

## Attributes Reference

Each result corresponds to one Redis cluster and exposes the same attributes as the [`scaleway_redis_cluster` resource](../resources/redis_cluster.md).
