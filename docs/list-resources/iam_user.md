---
page_title: "Scaleway: scaleway_iam_user"
subcategory: "IAM"
description: |-
  Lists Scaleway IAM Users.
---

# Resource: scaleway_iam_user

Lists Scaleway IAM Users.

For more information, see [the main documentation](https://www.scaleway.com/en/docs/iam/concepts/).

## Example Usage

```terraform
// List all users in an organization
list "scaleway_iam_user" "all" {
  provider = scaleway

  config {
    organization_id = "11111111-1111-1111-1111-111111111111"
  }
}
```
```terraform
// List users filtered by MFA status
list "scaleway_iam_user" "by_mfa" {
  provider = scaleway

  config {
    mfa = true
  }
}
```
```terraform
// List users filtered by tag
list "scaleway_iam_user" "by_tag" {
  provider = scaleway

  config {
    tag = "production"
  }
}
```
```terraform
// List users filtered by user IDs
list "scaleway_iam_user" "by_user_ids" {
  provider = scaleway

  config {
    user_ids = [
      "11111111-1111-1111-1111-111111111111",
      "22222222-2222-2222-2222-222222222222",
    ]
  }
}
```

## Argument Reference

The following arguments can be specified in the `config` block:

- `organization_id` - (Optional) Organization ID to filter for. If not specified, the provider default organization is used.
- `tag` - (Optional) Filter by tags containing a given string.
- `mfa` - (Optional) Filter by MFA status.
- `type` - (Optional) Filter by user type.
- `user_ids` - (Optional) Filter users by user IDs.

## Attributes Reference

In addition to the arguments above, the following attributes are exported for each User:

- `id` - The ID of the user.
- `organization_id` - The organization ID the user belongs to.
- `email` - The email of the user.
- `username` - The member's username.
- `tags` - The tags associated with the user.
- `first_name` - The member's first name.
- `last_name` - The member's last name.
- `phone_number` - The member's phone number.
- `locale` - The member's locale.
- `created_at` - The date and time of the creation of the iam user.
- `updated_at` - The date and time of the last update of the iam user.
- `deletable` - Whether or not the iam user is editable.
- `last_login_at` - The date and time of last login of the iam user.
- `type` - The type of the iam user.
- `status` - The status of user invitation.
- `mfa` - Whether or not the MFA is enabled.
- `account_root_user_id` - The ID of the account root user associated with the iam user.
- `locked` - Defines whether the user is locked.
