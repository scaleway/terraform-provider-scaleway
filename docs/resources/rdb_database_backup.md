---
page_title: "Scaleway: scaleway_rdb_database_backup"
description: |-
Manages Scaleway RDB Database Backup.
---

# scaleway_rdb_database_backup

Creates and manages Scaleway RDB database backup.
For more information, see [the documentation](https://developers.scaleway.com/en/products/rdb/api).

## Examples

### Basic

```hcl
resource scaleway_rdb_database_backup "main" {
  instance_id = data.scaleway_rdb_instance.main.id
  database_name = data.scaleway_rdb_database.main.name
}
```

### With expiration

```hcl
resource scaleway_rdb_database_backup "main" {
  instance_id = data.scaleway_rdb_instance.main.id
  database_name = data.scaleway_rdb_database.main.name
  expires_at = "2022-06-16T07:48:44Z"
}
```

## Arguments Reference

The following arguments are supported:

- `instance_id` - (Required) UUID of the instance where the database to backup is.

~> **Important:** Updates to `instance_id` will recreate the Backup.

- `name` - (Required) Name of the database (e.g. `my-database`).

- `expires_at` (Optional) Expiration date (Format ISO 8601).

~> **Important:** `expires_at` cannot be removed after being set.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `size` - Size of the backup (in bytes).
- `instance_name` - Name of the instance of the backup.
- `created_at` - Creation date (Format ISO 8601).
- `updated_at` - Updated date (Format ISO 8601).


## Import

RDB Database can be imported using the `{region}/{id}`, e.g.

```bash
$ terraform import scaleway_rdb_database_backup.mybackup fr-par/11111111-1111-1111-1111-111111111111
```
