---
page_title: "Scaleway: scaleway_rdb_privilege"
description: |-
  Manages Scaleway RDB Database Privilege.
---

# scaleway_rdb_privilege

Create and manage Scaleway RDB database privilege.
For more information, see [the documentation](https://developers.scaleway.com/en/products/rdb/api/#user-and-permissions).

## Example usage


```hcl
resource "scaleway_rdb_privilege" "priv" {
  instance_id   = scaleway_rdb_instance.rdb.id
  user_name     = "my-db-user"
  database_name = "my-db-name"
  permission    = "all"

  depends_on = [scaleway_rdb_user.main, scaleway_rdb_database.main]
}

resource "scaleway_rdb_user" "main" {
  instance_id = scaleway_rdb_instance.pgsql.id
  name        = "foobar"
  password    = "thiZ_is_v&ry_s3cret"
  is_admin    = false
}

resource "scaleway_rdb_database" "main" {
  instance_id = scaleway_rdb_instance.pgsql.id
  name        = "foobar"
}
```

## Argument Reference

The following arguments are supported:

- `instance_id` - (Required) UUID of the instance where to create the database.

- `user_name` - (Required) Name of the user (e.g. `my-db-user`).

- `database_name` - (Required) Name of the database (e.g. `my-db-name`).

- `permission` - (Required) Permission to set. Valid values are `readonly`, `readwrite`, `all`, `custom` and `none`.

## Attributes Reference

In addition to all above arguments, the following attribute is exported:

- `region` - The Scaleway region this resource resides in.

## Import

The user privileges can be imported using the `{region}/{instance_id}/{database_name}/{user_name}`, e.g.

```bash
$ terraform import scaleway_rdb_privilege.o fr-par/11111111-1111-1111-1111-111111111111/database_name/foo
```
