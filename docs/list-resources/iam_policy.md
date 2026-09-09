---
page_title: "Scaleway: scaleway_iam_policy"
subcategory: "IAM"
description: |-
  Lists Scaleway IAM Policies.
---

# Resource: scaleway_iam_policy

Lists Scaleway IAM Policies.

For more information, see [the main documentation](https://www.scaleway.com/en/docs/iam/concepts/).

## Example Usage

```terraform
// List all policies in an organization
list "scaleway_iam_policy" "all" {
  provider = scaleway

  config {
    organization_id = "11111111-1111-1111-1111-111111111111"
  }
}
```

```terraform
// List policies by policy IDs
list "scaleway_iam_policy" "by_ids" {
  provider = scaleway

  config {
    organization_id = "11111111-1111-1111-1111-111111111111"
    policy_ids      = ["11111111-1111-1111-1111-111111111111", "22222222-2222-2222-2222-222222222222"]
  }
}
```

```terraform
// List policies by tag
list "scaleway_iam_policy" "by_tag" {
  provider = scaleway

  config {
    organization_id = "11111111-1111-1111-1111-111111111111"
    tag             = "production"
  }
}
```

```terraform
// List policies by user IDs
list "scaleway_iam_policy" "by_user_ids" {
  provider = scaleway

  config {
    organization_id = "11111111-1111-1111-1111-111111111111"
    user_ids        = ["11111111-1111-1111-1111-111111111111"]
  }
}
```


## Argument Reference

The following arguments can be specified in the `config` block:

- `organization_id` - (Optional) Organization ID to filter for. If not specified, the provider default organization is used.
- `tag` - (Optional) Filter by tags containing a given string.
- `editable` - (Optional) Filter by editable status.
- `policy_ids` - (Optional) Filter policies by policy IDs.
- `user_ids` - (Optional) Filter policies by user IDs.
- `group_ids` - (Optional) Filter policies by group IDs.
- `application_ids` - (Optional) Filter policies by application IDs.
- `no_principal` - (Optional) Filter by policies with no principal.
- `policy_name` - (Optional) Filter by policy name.

## Attributes Reference

In addition to the arguments above, the following attributes are exported for each Policy:

- `id` - The ID of the policy.
- `organization_id` - The organization ID the policy belongs to.
- `name` - The name of the policy.
- `description` - The description of the policy.
- `created_at` - The date and time of the creation of the policy.
- `updated_at` - The date and time of the last update of the policy.
- `editable` - Whether or not the policy is editable.
- `tags` - The tags associated with the policy.
- `user_id` - The user ID the policy is attached to.
- `group_id` - The group ID the policy is attached to.
- `application_id` - The application ID the policy is attached to.
- `no_principal` - Whether the policy is not attached to any principal.
- `rule` - The rules of the policy (empty in list resource).
