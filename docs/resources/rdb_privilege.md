---
page_title: "Scaleway: scaleway_rdb_privilege"
description: |-
  Manages Scaleway RDB Database Privilege.
---

# scaleway_rdb_privilege

Create and manage Scaleway RDB database privilege.
For more information, see [the documentation](https://developers.scaleway.com/en/products/rdb/api).

## Example usage


```hcl
resource "scaleway_rdb_privilege" "priv" {
  instance_id   = scaleway_rdb_instance.rdb.id
  user_name     = "my-db-user"
  database_name = "my-db-name"
  permission    = "all"
}
```

## Argument Reference

The following arguments are supported:

- `instance_id` - (Required) UUID of the instance where to create the database.

- `user_name` - (Required) Name of the user (e.g. `my-db-user`).

- `database_name` - (Required) Name of the database (e.g. `my-db-name`).

- `permission` - (Required) Permission to set. Valid values are `readonly`, `readwrite`, `all`, `custom` and `none`).

## Attributes Reference

- `instance_id` - See Argument Reference above.

- `user_name` - See Argument Reference above.

- `database_name` - See Argument Reference above.

- `permission` - See Argument Reference above.
