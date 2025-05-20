---
subcategory: "IAM"
page_title: "Scaleway: scaleway_iam_user"
---

# Resource: scaleway_iam_user

Creates and manages Scaleway IAM Users.
For more information, see the [API documentation](https://www.scaleway.com/en/developers/api/iam/#path-users-list-users-of-an-organization).

## Example Usage

### Guest user

```terraform
resource "scaleway_iam_user" "guest" {
  email = "foo@test.com"
  tags  = ["test-tag"]
}
```

### Member user

```terraform
resource "scaleway_iam_user" "member" {
  email      = "foo@test.com"
  tags       = ["test-tag"]
  username   = "foo"
  first_name = "Foo"
  last_name  = "Bar"
}
```

When `username` is set, the user is created as a [Member](https://www.scaleway.com/en/docs/iam/concepts/#member). Otherwise, it is created as a [Guest](https://www.scaleway.com/en/docs/iam/concepts/#guest).

### Multiple users

```terraform
locals {
  users = toset([
    "test@test.com",
    "test2@test.com"
  ])
}

resource "scaleway_iam_user" "user" {
  for_each = local.users
  email    = each.key
}
```

## Argument Reference

- `organization_id` - (Defaults to [provider](../index.md#organization_id) `organization_id`) The ID of the organization the user is associated with.

- `email` - (Required) The email of the IAM user. For Guest users, this argument is not editable.

- `tags` - (Optional) The tags associated with the user.

- `username` - (Optional) The username of the IAM user. When it is set, the user is created as a Member. When it is not set, the user is created as a Guest and the username is set as equal to the email.

- `password` - The password for first access.

- `send_password_email` - Whether or not to send an email containing the password for first access.

- `send_welcome_email` - Whether or not to send a welcome email that includes onboarding information.

- `first_name` - The user's first name.

- `last_name` - The user's last name.

- `phone_number` - The user's phone number.

- `locale` - The user's locale (e.g., en_US).

Important: When creating a Guest user, all arguments are ignored, except for `organization_id`, `email` and `tags`.

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
