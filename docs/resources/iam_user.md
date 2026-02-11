---
subcategory: "IAM"
page_title: "Scaleway: scaleway_iam_user"
---

# Resource: scaleway_iam_user

Creates and manages Scaleway IAM [Users](https://www.scaleway.com/en/docs/iam/concepts/#member).
For more information, see the [API documentation](https://www.scaleway.com/en/developers/api/iam/#path-users-list-users-of-an-organization).



## Example Usage

```terraform
### Basic IAM user creation

resource "scaleway_iam_user" "user" {
  email      = "foo@test.com"
  tags       = ["test-tag"]
  username   = "foo"
  first_name = "Foo"
  last_name  = "Bar"
}
```

```terraform
### Multiple IAM user creation

locals {
  users = [
    {
      email    = "test@test.com"
      username = "test"
    },
    {
      email    = "test2@test.com"
      username = "test2"
    }
  ]
}

resource "scaleway_iam_user" "users" {
  count    = length(local.users)
  email    = local.users[count.index].email
  username = local.users[count.index].username
}
```

```terraform
### Creating a user using a Write Only password (not stored in state)

## Generate an ephemeral password (not stored in the state)
ephemeral "random_password" "db_password" {
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

resource "scaleway_iam_user" "password_wo_user" {
  email               = "user@example.com"
  username            = "testuser"
  password_wo         = ephemeral.random_password.db_password.result
  password_wo_version = 1
}
```



## Argument Reference

- `organization_id` - (Defaults to [provider](../index.md#organization_id) `organization_id`) The ID of the organization the user is associated with.

- `email` - (Required) The email of the IAM user. For Guest users, this argument is not editable.

- `tags` - (Optional) The tags associated with the user.

- `username` - (Required) The username of the IAM user.

- `password` - The password for first access. Only one of `password` or `password_wo` should be specified.

- `password_wo` - (Optional) The password for first access in [write-only](https://developer.hashicorp.com/terraform/language/manage-sensitive-data/write-only) mode. Only one of `password` or `password_wo` should be specified. `password_wo` will not be set in the Terraform state. To update the `password_wo`, you must also update the `password_wo_version`.

- `password_wo_version` - (Optional) The version of the [write-only](https://developer.hashicorp.com/terraform/language/manage-sensitive-data/write-only) password. To update the `password_wo`, you must also update the `password_wo_version`.

- `send_password_email` - Whether or not to send an email containing the password for first access.

- `send_welcome_email` - Whether or not to send a welcome email that includes onboarding information.

- `first_name` - The user's first name.

- `last_name` - The user's last name.

- `phone_number` - The user's phone number.

- `locale` - The user's locale (e.g., en_US).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the user (UUID format).
- `created_at` - The date and time of the creation of the IAM user.
- `updated_at` - The date and time of the last update of the IAM user.
- `deletable` - Whether the IAM user is deletable.
- `last_login_at` - The date of the last login.
- `type` - The type of user. Check the possible values in the [API doc](https://www.scaleway.com/en/developers/api/iam/#path-users-get-a-given-user).
- `status` - The status of user invitation. Check the possible values in the [API doc](https://www.scaleway.com/en/developers/api/iam/#path-users-get-a-given-user).
- `mfa` - Whether the MFA is enabled.
- `account_root_user_id` - The ID of the account root user associated with the user.
- `locked` - Whether the user is locked.

## Import

IAM users can be imported using the `{id}`, e.g.

```bash
terraform import scaleway_iam_user.basic 11111111-1111-1111-1111-111111111111
```
