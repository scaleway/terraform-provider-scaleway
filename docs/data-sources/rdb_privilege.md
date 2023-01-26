---
layout: "scaleway"
page_title: "Scaleway: scaleway_rdb_privilege"
description: |-
  Gets information about the privilege on RDB database.
---

# scaleway_rdb_privilege

Gets information about the privilege on a RDB database.
For more information, see [the documentation](https://developers.scaleway.com/en/products/rdb/api/#user-and-permissions).

## Example Usage

```hcl
# Get the database privilege for the user "my-user" on the database "my-database" hosted on instance id fr-par/11111111-1111-1111-1111-111111111111
data "scaleway_rdb_privilege" "main" {
  instance_id   = "fr-par/11111111-1111-111111111111"
  user_name     = "my-user"
  database_name = "my-database"
}
```

## Argument Reference

- `instance_id` - (Required) The RDB instance ID.

- `user_name` - (Required) The user name.

- `database_name` - (Required) The database name.

## Attributes Reference

In addition to all above arguments, the following attribute is exported:

- `id` - The resource ID
- `region` - The Scaleway region this bucket resides in.
- `permission` - The permission for this user on the database. Possible values are `readonly`, `readwrite`, `all`
  , `custom` and `none`.
