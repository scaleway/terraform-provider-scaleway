The [`scaleway_account_project`](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/resources/account_project) resource allows you to create and manage the Projects of a Scaleway Organization.

Refer to the Organizations and Projects [documentation](https://www.scaleway.com/en/docs/organizations-and-projects/) and [API documentation](https://www.scaleway.com/en/developers/api/account/project-api/) for more information.

If the project still contains resources at destroy time, the provider deletes the project along with all the resources it contains, including resources not managed by Terraform. Deleting a project is permanent and cannot be undone.
