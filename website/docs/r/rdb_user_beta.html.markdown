---
layout: "scaleway"
page_title: "Scaleway: scaleway_rdb_user_beta"
description: |-
  Manages Scaleway Database Users.
---

# scaleway_rdb_user_beta

-> **Note:** This terraform resource is flagged beta and might include breaking change in future releases.

Creates and manages Scaleway Database Users. For more information, see [the documentation](https://developers.scaleway.com/en/products/rdb/api).

## Examples

### Basic

```hcl
resource scaleway_rdb_user_beta db_admin {
  	instance_id = scaleway_rdb_instance_beta.main.id
  	name = "titi"
  	password = "R34lP4sSw#Rd"
  	is_admin = true
}
```

## Arguments Reference

The following arguments are supported:

- `instance_id` - (Required) The instance on which to create the user.

~> **Important:** Updates to `instance_id` will recreate the Database User.

- `name` - (Required) Database User name.

~> **Important:** Updates to `name` will recreate the Database User.

- `password` - (Required) Database User password.

- `is_admin` - (Optional) Grand admin permissions to the Database User.

## Import

Database User can be imported using `{region}/{instance_id}/{name}`, e.g.
```bash
$ terraform import scaleway_rdb_user_beta.admin fr-par/11111111-1111-1111-1111-111111111111/admin
```
