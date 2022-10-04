---
page_title: "Scaleway: scaleway_iam_user"
description: |-
Get information on an existing IAM user.
---

# scaleway_iam_user

| WARNING: This resource is in beta version. If your are in the beta group, please set the variable `SCW_ENABLE_BETA=true` in your `env` in order to use this resource. |
|------------------------------------------------------------------------------------------------------------------------------------------------------------------|

Use this data source to get information on an existing IAM user based on its ID or email address.
For more information, see [the documentation](https://developers.prd.frt.internal.scaleway.com/en/products/iam/api/v1alpha1/#users-06bdcf).

## Example Usage

```hcl
# Get info by user id
data "scaleway_iam_user" "by_id" {
  user_id = "11111111-1111-1111-1111-111111111111"
  organization_id = "11111111-1111-1111-1111-111111111111"
}
# Get info by email address
data "scaleway_iam_user" "by_email" {
  email = "foo@bar.com"
  organization_id = "11111111-1111-1111-1111-111111111111"
}
```

## Argument Reference

- `email` - (Optional) The eamil address of the IAM user. Only one of the `email` and `user_id` should be specified.
- `user_id` - (Optional) The ID of the IAM user. Only one of the `email` and `user_id` should be specified.
- `organization_id` - (Required) The organization ID the IAM group is associated with. For now, it is necessary to explicitly provide the `organization_id` in the datasource.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the IAM user.
- `created_at` - The date and time of the creation of the IAM user.
- `updated_at` - The date and time of the last update of the IAM user.
- `deletable` - The deletion status of the IAM user. Owner user cannot be deleted
- `last_login_at` - The last login date of the IAM user
- `type` - The type of the IAM user. Possible values are unknown_type, guest and owner. The default value is unknown_type.
- `two_factor_enabled` - The 2FA status of the IAM user
- `status` - The invitation status of the IAM user. Possible values are unknown_status, invitation_pending and activated. The default value is unknown_status.
