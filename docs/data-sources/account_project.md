---
subcategory: "Account"
page_title: "Scaleway: scaleway_account_project"
---

# scaleway_account_project

The `scaleway_account_project` data source is used to retrieve information about a Scaleway project.

Refer to the Organizations and Projects [documentation](https://www.scaleway.com/en/docs/organizations-and-projects/) and [API documentation](https://www.scaleway.com/en/developers/api/account/project-api/) for more information.


## Retrieve a Scaleway Project

The following commands allow you to:

- retrieve a Project by its name
- retrieve a Project by its ID
- retrieve the default project of an Organization

```hcl
# Get info by name
data scaleway_account_project "by_name" {
  name            = "myproject"
  organization_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
}
# Get default project
data scaleway_account_project "by_name" {
  name = "default"
}
# Get info by ID
data scaleway_account_project "by_id" {
  project_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
}
```

## Argument Reference

This section lists the arguments that you can provide to the `scaleway_account_project` data source to filter and retrieve the desired project. Each argument has a specific purpose:

- `name` - (Optional) The name of the Project.
  Only one of the `name` and `project_id` should be specified.

- `project_id` - (Optional) The unique identifier of the Project.
  Only one of the `name` and `project_id` should be specified.

- `organization_id` - (Optional) The unique identifier of the Organization with which the Project is associated.

  If no default `organization_id` is set, one must be set explicitly in this datasource

## Attribute reference

The `scaleway_account_project` data source exports certain attributes once the account information is retrieved. These attributes can be referenced in other parts of your Terraform configuration. The exported attributes come from the `account_project` [resource](../resources/account_project.md).
