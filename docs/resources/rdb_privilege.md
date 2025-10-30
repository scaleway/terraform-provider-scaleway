---
subcategory: "Databases"
page_title: "Scaleway: scaleway_rdb_privilege"
---

# Resource: scaleway_rdb_privilege

Create and manage Scaleway database privileges.
For more information refer to the [API documentation](https://www.scaleway.com/en/developers/api/managed-database-postgre-mysql/#user-and-permissions).

## Example Usage

```terraform
resource "scaleway_rdb_instance" "main" {
  name           = "rdb"
  node_type      = "DB-DEV-S"
  engine         = "PostgreSQL-11"
  is_ha_cluster  = true
  disable_backup = true
  user_name      = "my_initial_user"
  password       = "thiZ_is_v&ry_s3cret"
}

resource "scaleway_rdb_database" "main" {
  instance_id = scaleway_rdb_instance.main.id
  name        = "database"
}

resource "scaleway_rdb_user" "main" {
  instance_id = scaleway_rdb_instance.main.id
  name        = "my-db-user"
  password    = "thiZ_is_v&ry_s3cret"
  is_admin    = false
}

resource "scaleway_rdb_privilege" "main" {
  instance_id   = scaleway_rdb_instance.main.id
  user_name     = scaleway_rdb_user.main.name
  database_name = scaleway_rdb_database.main.name
  permission    = "all"
}
```

## Argument Reference

The following arguments are supported:

- `instance_id` - (Required) UUID of the Database Instance.

- `user_name` - (Required) Name of the user (e.g. `my-db-user`).

- `database_name` - (Required) Name of the database (e.g. `my-db-name`).

- `permission` - (Required) Desired permission level. Valid values are `readonly`, `readwrite`, `all`, `custom` and `none`.

- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions) in which the resource exists.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the user privileges, which is of the form `{region}/{instance_id}/{database_name}/{user_name}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111/database_name/foo`

- `effective_permission` - The actual permission currently set in Scaleway. May differ from `permission` after database schema changes (new tables, views, or sequences created).

- `permission_status` - Permission synchronization status. Possible values:
    - `synced`: The effective permission matches the desired permission
    - `drifted`: The effective permission differs from the desired permission (requires `terraform apply` to resync)

## Permission Drift Management

### Understanding Permission Drift

When you configure a privilege (e.g., `readwrite`), Scaleway applies it to **database objects that exist at that moment**. If new tables, views, or sequences are created later, they won't automatically inherit these permissions. In that case, the API may return `custom`.

**Example:**

```terraform
resource "scaleway_rdb_privilege" "app" {
  instance_id   = scaleway_rdb_instance.main.id
  user_name     = "app_user"
  database_name = "mydb"
  permission    = "readwrite"
  
  # Later, after new objects are created externally:
  # effective_permission = "custom"  (computed)
  # permission_status    = "drifted" (computed)
}
```

### Handling Permission Drift

Run `terraform apply` to reapply the configured permission to all objects (existing and new):

```bash
terraform apply
```

The plan will typically show:

```diff
~ resource "scaleway_rdb_privilege" "app" {
    ~ permission = "custom" -> "readwrite"
}
```

## Import

The user privileges can be imported using the `{region}/{instance_id}/{database_name}/{user_name}`, e.g.

```bash
terraform import scaleway_rdb_privilege.o fr-par/11111111-1111-1111-1111-111111111111/database_name/foo
```
