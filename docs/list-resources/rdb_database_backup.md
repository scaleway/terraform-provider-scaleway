---
page_title: "Scaleway: scaleway_rdb_database_backup"
subcategory: "Databases"
description: |-
  Lists Scaleway RDB database backups for Database Instances across regions and projects.
---

# Resource: scaleway_rdb_database_backup



For more information, see the [product documentation](https://www.scaleway.com/en/docs/managed-databases/postgresql-and-mysql/).


## Example Usage

```terraform
# List database backups on a specific RDB instance
list "scaleway_rdb_database_backup" "by_instance" {
  provider = scaleway

  config {
    regions      = ["fr-par"]
    project_ids  = ["11111111-1111-1111-1111-111111111111"]
    instance_ids = ["fr-par/22222222-2222-2222-2222-222222222222"]
  }
}
```

```terraform
# List database backups filtered by name
list "scaleway_rdb_database_backup" "by_name" {
  provider = scaleway

  config {
    regions      = ["fr-par"]
    project_ids  = ["11111111-1111-1111-1111-111111111111"]
    instance_ids = ["*"]
    name         = "my-backup"
  }
}
```

```terraform
# List database backups on all RDB instances in a region and project
list "scaleway_rdb_database_backup" "all_instances" {
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

- `regions` - (Optional) Regions of the RDB Database Instance to filter on. Use `["*"]` to list from all regions. If omitted, the provider default region is used.
- `project_ids` - (Optional) Project IDs to filter for. Use `["*"]` to list across all projects. If omitted, the provider default project is used.
- `instance_ids` - (Required) Database Instance IDs to list backups from. Use `["*"]` only to include all instances in each selected region and project. Otherwise each value must be a regional ID (`region/uuid`) or a bare instance UUID.
- `name` - (Optional) Name of the database backup to filter on.

## Attributes Reference

Each result corresponds to one RDB database backup and exposes the same attributes as the [`scaleway_rdb_database_backup` resource](../resources/rdb_database_backup.md).
