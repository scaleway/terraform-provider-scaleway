---
subcategory: "IAM"
page_title: "Scaleway: scaleway_iam_group_membership"
---

# Resource: scaleway_iam_group_membership

Add members to an IAM group.
For more information refer to the [IAM API documentation](https://www.scaleway.com/en/developers/api/iam/#groups-f592eb).

## Example Usage

### Application Membership

```terraform
resource "scaleway_iam_group" "group" {
  name                = "my_group"
  external_membership = true
}

resource "scaleway_iam_application" "app" {
  name = "my_app"
}

resource "scaleway_iam_group_membership" "member" {
  group_id       = scaleway_iam_group.group.id
  application_ids = [scaleway_iam_application.app.id]
}
```

### Users membership

```terraform
locals {
  users = toset([
    "user1@mail.com",
    "user2@mail.com"
  ])
}

data "scaleway_iam_user" "users" {
  for_each = local.users
  email    = each.value
}

resource "scaleway_iam_group" "group" {
  name                = "my_group"
  external_membership = true
}

resource "scaleway_iam_group_membership" "members" {
  for_each = data.scaleway_iam_user.users
  group_id = scaleway_iam_group.group.id
  user_ids = [each.value.id]
}
```

## Argument Reference

- `group_id` - (Required) ID of the group to add members to.

- `application_ids` - (Optional) The IDs of the applications that will be added to the group.

- `user_ids` - (Optional) The IDs of the users that will be added to the group

  -> **Note** You must specify at least one: `application_ids` and/or `user_ids`.

## Attributes Reference

No additional attributes are exported.

## Import

IAM group memberships can be imported using the following format:

- For user: `{group_id}/user:userID,application:applicationID,...`

```bash
terraform import scaleway_iam_group_membership.members 11111111-1111-1111-1111-111111111111/user:11111111-1111-1111-1111-111111111111,application:11111111-1111-1111-1111-111111111111
```
