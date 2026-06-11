---
page_title: "Scaleway: scaleway_iam_group"
subcategory: "IAM"
description: |-
  Lists Scaleway IAM Groups.
---

# Resource: scaleway_iam_group

Lists Scaleway IAM Groups.

For more information, see [the main documentation](https://www.scaleway.com/en/docs/iam/concepts/).

## Example Usage

```terraform
# List all groups in an organization
list "scaleway_iam_group" "all" {
  provider = scaleway

  config {
    organization_id = "11111111-1111-1111-1111-111111111111"
  }
}
```

```terraform
# List groups filtered by application IDs
list "scaleway_iam_group" "by_application" {
  provider = scaleway

  config {
    application_ids = [
      "11111111-1111-1111-1111-111111111111",
      "22222222-2222-2222-2222-222222222222",
    ]
  }
}
```

```terraform
# List groups filtered by name
list "scaleway_iam_group" "by_name" {
  provider = scaleway

  config {
    name = "my-group"
  }
}
```

```terraform
# List groups filtered by tag
list "scaleway_iam_group" "by_tag" {
  provider = scaleway

  config {
    tag = "production"
  }
}
```

```terraform
# List groups filtered by user IDs
list "scaleway_iam_group" "by_user" {
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
- `name` - (Optional) Name of the group to filter for.
- `tag` - (Optional) Tag to filter for.
- `user_ids` - (Optional) Filter groups by user IDs.
- `application_ids` - (Optional) Filter groups by application IDs.

## Attributes Reference

In addition to the arguments above, the following attributes are exported for each Group:

- `id` - The ID of the group.
- `name` - The name of the group.
- `description` - The description of the group.
- `created_at` - The date and time of the creation of the group.
- `updated_at` - The date and time of the last update of the group.
- `organization_id` - The organization ID the group belongs to.
- `tags` - The tags associated with the group.
- `user_ids` - List of IDs of the users attached to the group.
- `application_ids` - List of IDs of the applications attached to the group.
- `editable` - Defines whether or not the group is editable.
- `deletable` - Defines whether or not the group is deletable.
- `managed` - Defines whether or not the group is managed.
