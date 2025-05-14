---
subcategory: "Account"
page_title: "Scaleway: scaleway_account_projects"
---

# scaleway_account_projects

The `scaleway_account_projects` data source is used to list all Scaleway projects in an Organization.

Refer to the Organizations and Projects [documentation](https://www.scaleway.com/en/docs/organizations-and-projects/) and [API documentation](https://www.scaleway.com/en/developers/api/account/project-api/) for more information.


## Retrieve a Scaleway Projects

The following commands allow you to:

- retrieve all Projects in an Organization

```hcl
# Get all Projects in an Organization
data scaleway_account_projects "all" {
  organization_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
}
```

## Example Usage

### Deploy an SSH key in all your organization's projects

```hcl
data scaleway_account_projects "all" {}

resource "scaleway_account_ssh_key" "main" {
  name       = "main"
  public_key = local.public_key
  count      = length(data.scaleway_account_projects.all.projects)
  project_id = data.scaleway_account_projects.all.projects[count.index].id
}
```

## Argument Reference

- `organization_id` - (Optional) The unique identifier of the Organization with which the Projects are associated.
  If no default `organization_id` is set, one must be set explicitly in this datasource


## Attribute reference

The `scaleway_account_projects` data source exports the following attributes:

- `projects` - (Computed) A list of projects. Each project has the following attributes:
  - `id` - (Computed) The unique identifier of the project.
  - `name` - (Computed) The name of the project.
  - `organization_id` - (Computed) The unique identifier of the organization with which the project is associated.
  - `created_at` - (Computed) The date and time when the project was created.
  - `updated_at` - (Computed) The date and time when the project was updated.
  - `description` - (Computed) The description of the project.
