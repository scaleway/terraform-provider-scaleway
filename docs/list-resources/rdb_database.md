---
page_title: "Scaleway: scaleway_rdb_database"
subcategory: "Databases"
description: |-
  Lists Scaleway RDB databases for Database Instances across regions and projects.
---

# Resource: scaleway_rdb_database



For more information, see the [product documentation](https://www.scaleway.com/en/docs/managed-databases/postgresql-and-mysql/).


## Example Usage

```terraform
# List databases on a specific RDB instance
list "scaleway_rdb_database" "by_instance" {
  provider = scaleway

  config {
    regions      = ["fr-par"]
    project_ids  = ["11111111-1111-1111-1111-111111111111"]
    instance_ids = ["fr-par/22222222-2222-2222-2222-222222222222"]
  }
}
```

```terraform
# List databases filtered by name on all instances in scope
list "scaleway_rdb_database" "by_name" {
  provider = scaleway

  config {
    regions      = ["fr-par"]
    project_ids  = ["11111111-1111-1111-1111-111111111111"]
    instance_ids = ["*"]
    name         = "mydb"
  }
}
```

```terraform
# List databases on all RDB instances in a region and project
list "scaleway_rdb_database" "all_instances" {
  provider = scaleway

  config {
    regions      = ["fr-par"]
    project_ids  = ["11111111-1111-1111-1111-111111111111"]
    instance_ids = ["*"]
  }
}
```



## Argument Reference

The following arguments can be specified in the `config` block:

- `regions` - (Optional) Regions of the Database Instance to filter on. Use `["*"]` to list from all regions. If omitted, the provider default region is used.
- `project_ids` - (Optional) Project IDs to filter for. Use `["*"]` to list across all projects. If omitted, the provider default project is used.
- `instance_ids` - (Required) Database Instance IDs to list databases from. Use `["*"]` only to include all instances in each selected region and project. Otherwise each value must be a regional ID (`region/uuid`) or a bare instance UUID.
- `name` - (Optional) Name of the database to filter on.
- `managed` - (Optional) When set, only databases with this managed flag are returned.
- `owner` - (Optional) Owner user name to filter on.
- `organization_id` - (Optional) Organization ID of the Database Instance to filter on when listing instances (wildcard `instance_ids`).

## Attributes Reference

Each result corresponds to one RDB database and exposes the same attributes as the [`scaleway_rdb_database` resource](../resources/rdb_database.md).
