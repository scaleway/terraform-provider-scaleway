---
subcategory: "Databases"
page_title: "Scaleway: scaleway_rdb_user"
---

# Resource: scaleway_rdb_user

Creates and manages database users.
For more information refer to the [API documentation](https://www.scaleway.com/en/developers/api/managed-database-postgre-mysql/).



## Example Usage

```terraform
### Basic user creation

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
  length      = 20
  special     = true
  upper       = true
  lower       = true
  numeric     = true
  min_upper   = 1
  min_lower   = 1
  min_numeric = 1
  min_special = 1
  # Exclude characters that might cause issues in some contexts
  override_special = "!@#$%^&*()_+-=[]{}|;:,.<>?"
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

- `password` - (Optional) database user password. The password must meet the following requirements based on ISO27001 standards:
    - **Length**: 8-128 characters
    - **Character types required**:
        - At least 1 lowercase letter (a-z)
        - At least 1 uppercase letter (A-Z)
        - At least 1 digit (0-9)
        - At least 1 special character (!@#$%^&*()_+-=[]{}|;:,.<>?)

    For secure password generation, consider using the `random_password` resource with appropriate parameters.

- `password_wo` - (Optional) Database user password in [write-only](https://developer.hashicorp.com/terraform/language/manage-sensitive-data/write-only) mode. Only one of `password` or `password_wo` should be specified. `password_wo` will not be set in the Terraform state. To update the `password_wo`, you must also update the `password_wo_version`.

- `password_wo_version` - (Optional) The version of the [write-only](https://developer.hashicorp.com/terraform/language/manage-sensitive-data/write-only) password. To update the `password_wo`, you must also update the `password_wo_version`.

- `is_admin` - (Optional) Grant admin permissions to the database user.

- `region` - The Scaleway region this resource resides in.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the user, which is of the form `{region}/{instance_id}/{user_name}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111/admin`

## Import

Database users can be imported using `{region}/{instance_id}/{user_name}`, e.g.

```bash
terraform import scaleway_rdb_user.admin fr-par/11111111-1111-1111-1111-111111111111/admin
```
