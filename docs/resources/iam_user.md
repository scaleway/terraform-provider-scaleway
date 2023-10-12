---
subcategory: "IAM"
page_title: "Scaleway: scaleway_iam_user"
---

# scaleway_iam_user

Creates and manages Scaleway IAM Users.
For more information, see [the documentation](https://www.scaleway.com/en/developers/api/iam/#path-users-list-users-of-an-organization).

## Examples

### Basic

```hcl
resource "scaleway_iam_user" "basic" {
  email            = "test@test.com"
}
```

## Argument Reference

- `email` - (Required) The email of the IAM user.

- `organization_id` - (Defaults to [provider](../index.md#organization_d) `organization_id`) The ID of the organization the user is associated with.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the user (UUID format).
- `email` - The email of the user
- `created_at` - The date and time of the creation of the iam user.
- `updated_at` - The date and time of the last update of the iam user.
- `deletable` - Whether the iam user is deletable.
- `organization_id` - The ID of the organization the user.
- `last_login_at` - The date of the last login.
- `type` - The type of user. Check the possible values on
  our [sdk](https://github.com/scaleway/scaleway-sdk-go/blob/master/api/iam/v1alpha1/iam_sdk.go#L508-L515C2).
- `two_factor_enabled` - Deprecated, use "mfa" instead.
- `status` - The status of user invitation. Check the possible values on
  our [sdk](https://github.com/scaleway/scaleway-sdk-go/blob/master/api/iam/v1alpha1/iam_sdk.go#L475-L480).
- `mfa` - Whether the MFA is enabled.
- `account_root_user_id` - The ID of the account root user associated with the user.

## Import

IAM users can be imported using the `{id}`, e.g.

```bash
$ terraform import scaleway_iam_user.basic 11111111-1111-1111-1111-111111111111
```
