---
subcategory: "Databases"
page_title: "Scaleway: scaleway_rdb_database_backup"
---

# Resource: scaleway_rdb_database_backup

Creates and manages database backups.
For more information, refer to the [API documentation](https://www.scaleway.com/en/developers/api/managed-database-postgre-mysql/).

## Example Usage

### Basic

```terraform
resource "scaleway_rdb_instance" "main" {
  name           = "test-rdb"
  node_type      = "DB-DEV-S"
  engine         = "PostgreSQL-15"
  is_ha_cluster  = true
  disable_backup = true
  user_name      = "my_initial_user"
  password       = "thiZ_is_v&ry_s3cret"
}

resource "scaleway_rdb_database" "main" {
  instance_id = scaleway_rdb_instance.main.id
  name        = "database"
}

resource scaleway_rdb_database_backup "main" {
  instance_id   = scaleway_rdb_instance.main.id
  database_name = scaleway_rdb_database.main.name
}
```

### With expiration

```terraform
resource scaleway_rdb_database_backup "main" {
  instance_id   = data.scaleway_rdb_instance.main.id
  database_name = data.scaleway_rdb_database.main.name
  expires_at    = "2022-06-16T07:48:44Z"
}
```

## Argument Reference

The following arguments are supported:

- `instance_id` - (Required) UUID of the Database Instance.

~> **Important:** Updates to `instance_id` will recreate the backup.

- `name` - (Required) Name of the database (e.g. `my-database`).

- `expires_at` (Optional) Expiration date (Format ISO 8601).

~> **Important:** `expires_at` cannot be removed after being set.

- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions) in which the resource exists.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the backup, which is of the form `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111`
- `size` - Size of the backup (in bytes).
- `instance_name` - Name of the instance of the backup.
- `created_at` - Creation date (Format ISO 8601).
- `updated_at` - Updated date (Format ISO 8601).

## Import

Databases can be imported using the `{region}/{id}`, e.g.

```bash
terraform import scaleway_rdb_database_backup.mybackup fr-par/11111111-1111-1111-1111-111111111111
```
