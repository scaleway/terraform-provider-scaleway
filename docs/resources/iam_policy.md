---
page_title: "Scaleway: scaleway_iam_policy"
description: |-
Manages Scaleway IAM Policies.
---

# scaleway_iam_policy

| WARNING: This resource is in beta version. If your are in the beta group, please set the variable `SCW_ENABLE_BETA=true` in your `env` in order to use this resource. |
|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------|

Creates and manages Scaleway IAM Policies. For more information, see [the documentation](https://developers.scaleway.com/en/products/iam/api/v1alpha1/#policies-54b8a7).

## Example Usage

```hcl
resource "scaleway_iam_policy" "main" {
  name = "my policy"
  description = "a description"
  no_principal = true
  rule {
    project_ids = ["xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"]
    permission_set_names = ["AllProductsFullAccess"]
  }
}
```

## Arguments Reference

The following arguments are supported:

- `name` - .The name of the iam policy.
- `description` - The description of the iam policy.
- `organization_id` - (Defaults to [provider](../index.md#organization_d) `organization_id`) The ID of the organization the policy is associated with.
- `user_id` - ID of the User the policy will be linked to
- `group_id` - ID of the Group the policy will be linked to
- `application_id` - ID of the Application the policy will be linked to
- `no_principal` - If the policy doesn't apply to a principal.

~> **Important** Only one of `user_id`, `group_id`, `application_id` and `no_principal`  may be set.

- `rule` - List of rules in the policy.
    - `organization_id` - ID of organization scoped to the rule.
    - `project_ids` - List of project IDs scoped to the rule.

  ~> **Important** One of `organization_id` or `project_ids`  must be set per rule.

    - `permission_set_names` - Names of permission sets bound to the rule.

  **_TIP:_**  You can use the Scaleway CLI to list the permissions details. e.g:

```shell
  $ SCW_ENABLE_BETA=1 scw iam permission-set list
```

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `created_at` - The date and time of the creation of the policy.
- `updated_at` - The date and time of the last update of the policy.
- `editable` - Whether the policy is editable.

## Import

Policies can be imported using the `{id}`, e.g.

```bash
$ terraform import scaleway_iam_policy.main 11111111-1111-1111-1111-111111111111
```
