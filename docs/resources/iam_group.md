---
page_title: "Scaleway: scaleway_iam_group"
description: |-
Manages Scaleway IAM Groups.
---

# scaleway_iam_group

| WARNING: This resource is in beta version. If your are in the beta group, please set the variable `SCW_ENABLE_BETA=true` in your `env` in order to use this resource. |
|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------|

Creates and manages Scaleway IAM Groups.

## Examples

### Basic

```hcl
resource "scaleway_iam_group" "basic" {
  name = "iam_group_basic"
  description = "basic description"
  application_ids = []
  user_ids = []
}
```

### With applications

```hcl
resource "scaleway_iam_application" "app" {}

resource "scaleway_iam_group" "with_app" {
  name = "iam_group_with_app"
  application_ids = [
    scaleway_iam_application.app.id,
  ]
  user_ids = []
}
```

### With users

```hcl
resource "scaleway_iam_group" "with_users" {
  name = "iam_group_with_app"
  application_ids = []
  user_ids = [
    "11111111-1111-1111-1111-111111111111",
    "22222222-2222-2222-2222-222222222222",
  ]
}
```

## Argument Reference

- `name` - (Optional) The name of the IAM group.

- `description` - (Optional) The description of the IAM group.

- `application_ids` - (Required) The list of IDs of the applications attached to the group.

- `user_ids` - (Required) The list of IDs of the users attached to the group.

-> **Note:** Keep in mind that updating one of the fields `application_ids` or `user_ids` will have consequences on the
other so be sure to always specify the desired state of these IDs at every change, otherwise some could get overwritten.

## Import

IAM groups can be imported using the `{zone}/{id}`, e.g.

```bash
$ terraform import scaleway_iam_group.basic fr-par/11111111-1111-1111-1111-111111111111
```
