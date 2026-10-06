---
page_title: "Scaleway: scaleway_account_project"
subcategory: "Account"
description: |-
  Lists Scaleway Account Projects.
---

# Resource: scaleway_account_project

Lists Scaleway Account Projects.

For more information, see [the main documentation](https://www.scaleway.com/en/docs/iam/concepts/).

## Example Usage

```terraform
// List all projects in an organization
list "scaleway_account_project" "all" {
  provider = scaleway

  config {
    organization_id = "11111111-1111-1111-1111-111111111111"
  }
}
```

```terraform
// List projects filtered by name
list "scaleway_account_project" "by_name" {
  provider = scaleway

  config {
    name = "my-project"
  }
}
```

```terraform
// List projects filtered by project IDs
list "scaleway_account_project" "by_project_ids" {
  provider = scaleway

  config {
    project_ids = [
      "11111111-1111-1111-1111-111111111111",
      "22222222-2222-2222-2222-222222222222",
    ]
  }
}
```


## Argument Reference

The following arguments can be specified in the `config` block:

- `organization_id` - (Optional) Organization ID to filter for. If not specified, the provider default organization is used.
- `name` - (Optional) Filter by project name containing a given string.
- `project_ids` - (Optional) Filter projects by project IDs.

## Attributes Reference

In addition to the arguments above, the following attributes are exported for each Project:

- `id` - The ID of the project.
- `name` - The name of the project.
- `description` - Description of the project.
- `organization_id` - The organization ID the project belongs to.
- `created_at` - The date and time of the creation of the project (Format ISO 8601).
- `updated_at` - The date and time of the last update of the project (Format ISO 8601).
