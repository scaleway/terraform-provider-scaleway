---
page_title: "Scaleway: scaleway_rdb_instance"
description: |-
  Manages Scaleway Database Instances.
---

# scaleway_rdb_instance

Creates and manages Scaleway Database Instances.
For more information, see [the documentation](https://developers.scaleway.com/en/products/rdb/api).

## Examples

### Basic

```hcl
resource "scaleway_rdb_instance" "main" {
  name           = "test-rdb"
  node_type      = "DB-DEV-S"
  engine         = "PostgreSQL-11"
  is_ha_cluster  = true
  disable_backup = true
  user_name      = "my_initial_user"
  password       = "thiZ_is_v&ry_s3cret"
}

# with backup schedule
resource "scaleway_rdb_instance" "main" {
  name          = "test-rdb"
  node_type     = "DB-DEV-S"
  engine        = "PostgreSQL-11"
  is_ha_cluster = true
  user_name     = "my_initial_user"
  password      = "thiZ_is_v&ry_s3cret"
  
  disable_backup = true
  backup_schedule_frequency = 24 # every day
  backup_schedule_retention = 7  # keep it one week
}
```

## Arguments Reference

The following arguments are supported:

- `node_type` - (Required) The type of database instance you want to create (e.g. `db-dev-s`).

~> **Important:** Updates to `node_type` will upgrade the Database Instance to the desired `node_type` without any interruption. Keep in mind that you cannot downgrade a Database Instance.

- `engine` - (Required) Database Instance's engine version (e.g. `PostgreSQL-11`).

~> **Important:** Updates to `engine` will recreate the Database Instance.

- `volume_type` - (Optional, default to `lssd`) Type of volume where data are stored (`bssd` or `lssd`).

- `volume_size_in_gb` - (Optional) Volume size (in GB) when `volume_type` is set to `bssd`. Must be a multiple of 5000000000.

- `user_name` - (Optional) Identifier for the first user of the database instance.

~> **Important:** Updates to `user_name` will recreate the Database Instance.

- `password` - (Optional) Password for the first user of the database instance.

- `is_ha_cluster` - (Optional) Enable or disable high availability for the database instance.

~> **Important:** Updates to `is_ha_cluster` will recreate the Database Instance.

- `name` - (Optional) The name of the Database Instance.

- `disable_backup` - (Optional) Disable automated backup for the database instance.

- `backup_schedule_frequency` - (Optional) Backup schedule frequency in hours.

- `backup_schedule_retention` - (Optional) Backup schedule retention in days.

- `settings` - Map of engine settings to be set.

- `tags` - (Optional) The tags associated with the Database Instance.

- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions) in which the Database Instance should be created.

- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the Database Instance is associated with.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the Database Instance.
- `endpoint_ip` - The IP of the Database Instance.
- `endpoint_port` - The port of the Database Instance.
- `read_replicas` - List of read replicas of the database instance.
    - `ip` - IP of the replica.
    - `port` - Port of the replica.
    - `name` - Name of the replica.
- `certificate` - Certificate of the database instance.
- `organization_id` - The organization ID the Database Instance is associated with.

## Import

Database Instance can be imported using the `{region}/{id}`, e.g.

```bash
$ terraform import scaleway_rdb_instance.rdb01 fr-par/11111111-1111-1111-1111-111111111111
```
