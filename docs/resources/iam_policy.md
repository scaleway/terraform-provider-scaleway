---
subcategory: "IAM"
page_title: "Scaleway: scaleway_iam_policy"
---

# Resource: scaleway_iam_policy

Creates and manages Scaleway IAM Policies. For more information refer to the [IAM API documentation](https://www.scaleway.com/en/developers/api/iam/#path-policies-create-a-new-policy).

-> You can find a detailed list of all permission sets available at Scaleway in the permission sets [reference page](https://www.scaleway.com/en/docs/iam/reference-content/permission-sets/).

## Example Usage

### Create a policy for an organization's project

```terraform
provider scaleway {
  organization_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
}

data scaleway_account_project "default" {
  name = "default"
}

resource scaleway_iam_application "app" {
  name = "my app"
}

resource scaleway_iam_policy "object_read_only" {
  name = "my policy"
  description = "gives app readonly access to object storage in project"
  application_id = scaleway_iam_application.app.id
  rule {
    project_ids = [data.scaleway_account_project.default.id]
    permission_set_names = ["ObjectStorageReadOnly"]
  }
}
```

### Create a policy for all current and future projects in an organization

```terraform
resource scaleway_iam_application "app" {
  name = "my app"
}

resource scaleway_iam_policy "object_read_only" {
  name = "my policy"
  description = "gives app readonly access to object storage in project"
  application_id = scaleway_iam_application.app.id
  rule {
    organization_id = scaleway_iam_application.app.organization_id
    permission_set_names = ["ObjectStorageReadOnly"]
  }
}
```

### Create a permission for multiple users using a group

```terraform
locals {
  users = [
    "user1@mail.com",
    "user2@mail.com",
  ]
  project_name = "default"
}

data "scaleway_account_project" "project" {
  name = local.project_name
}

data "scaleway_iam_user" "users" {
  for_each = toset(local.users)
  email    = each.value
}

resource "scaleway_iam_group" "with_users" {
  name     = "developers"
  user_ids = [for user in data.scaleway_iam_user.users : user.id]
}

resource "scaleway_iam_policy" "iam_tf_storage_policy" {
  name     = "developers permissions"
  group_id = scaleway_iam_group.with_users.id
  rule {
    project_ids          = [data.scaleway_account_project.project.id]
    permission_set_names = ["InstancesReadOnly"]
  }
}
```

### Create a policy with a particular condition

IAM policy rule can use a condition to be applied.
The following variables are available:

- `request.ip`
- `request.user_agent`
- `request.time`

```terraform
resource "scaleway_iam_policy" "main" {
  name         = "tf_tests_policy_condition"
  no_principal = true
  rule {
    organization_id      = "%s"
    permission_set_names = ["AllProductsFullAccess"]
    condition = "request.user_agent == 'My User Agent'"
  }
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Optional) The name of the IAM policy.
- `description` - (Optional) The description of the IAM policy.
- `condition` - (Optional) The condition of the IAM policy.
- `tags` - (Optional) The tags associated with the IAM policy.
- `organization_id` - (Defaults to [provider](../index.md#organization_d) `organization_id`) The ID of the organization the policy is associated with.
- `user_id` - ID of the user the policy will be linked to
- `group_id` - ID of the group the policy will be linked to
- `application_id` - ID of the application the policy will be linked to
- `no_principal` - If the policy doesn't apply to a principal.

~> **Important** Only one of `user_id`, `group_id`, `application_id` and `no_principal` may be set.

- `rule` - List of rules in the policy.

    - `organization_id` - ID of organization scoped to the rule, this can be used to create a rule for all projects in an organization.
    - `project_ids` - List of project IDs scoped to the rule.

    ~> **Important** One `organization_id` or `project_ids` must be set per rule.

    - `permission_set_names` - Names of permission sets bind to the rule.

  **_TIP:_** You can use the Scaleway CLI to list the permissions details. e.g:

```shell
   scw IAM permission-set list
```

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the policy.
- `created_at` - The date and time of the creation of the policy.
- `updated_at` - The date and time of the last update of the policy.
- `editable` - Whether the policy is editable.

## Import

Policies can be imported using the `{id}`, e.g.

```bash
terraform import scaleway_iam_policy.main 11111111-1111-1111-1111-111111111111
```
