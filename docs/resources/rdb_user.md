---
subcategory: "Databases"
page_title: "Scaleway: scaleway_rdb_user"
---

# Resource: scaleway_rdb_user

Creates and manages database users.
For more information refer to [the API documentation](https://www.scaleway.com/en/developers/api/managed-database-postgre-mysql/).

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

resource "random_password" "db_password" {
  length  = 16
  special = true
}

resource "scaleway_rdb_user" "db_admin" {
  instance_id = scaleway_rdb_instance.main.id
  name        = "devtools"
  password    = random_password.db_password.result
  is_admin    = true
}
```

## Argument Reference

The following arguments are supported:

- `instance_id` - (Required) UUID of the Database Instance.

~> **Important:** Updates to `instance_id` will recreate the database user.

- `name` - (Required) database user name.

~> **Important:** Updates to `name` will recreate the database user.

- `password` - (Required) database user password.

- `is_admin` - (Optional) Grant admin permissions to the database user.

- `region` - The Scaleway region this resource resides in.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the user, which is of the form `{region}/{instance_id}/{user_name}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111/admin`

## Import

Database users can be imported using `{region}/{instance_id}/{user_name}`, e.g.

```bash
$ terraform import scaleway_rdb_user.admin fr-par/11111111-1111-1111-1111-111111111111/admin
```
