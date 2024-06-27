---
subcategory: "Account"
page_title: "Scaleway: scaleway_account_project"
---

# Resource: scaleway_account_project

The `scaleway_account_project` resource allows you to create and manage the Projects of a Scaleway Organization.

Refer to the Organizations and Projects [documentation](https://www.scaleway.com/en/docs/identity-and-access-management/organizations-and-projects/) and [API documentation](https://www.scaleway.com/en/developers/api/account/project-api/) for more information.

## Create a Scaleway Project

The following command allows you to create a project named `project`.

```hcl
resource "scaleway_account_project" "project" {
  name = "project"
}
```

## Use a project in provider configuration

If you want to use as default a project created in terraform you can use a temporary provider alias.
This project can then be used to configure your default provider.

```hcl
provider "scaleway" {
  alias = "tmp"
}

resource scaleway_account_project "project" {
  provider = scaleway.tmp
  name = "my_project"
}

provider "scaleway" {
  project_id = scaleway_account_project.project.id
}

resource "scaleway_instance_server" "server" { // Will use scaleway_account_project.project
  image = "ubuntu_jammy"
  type  = "PRO2-XXS"
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Optional) The name of the Project.
- `description` - (Optional) The description of the Project.
- `organization_id` - (Optional. Defaults to [provider](../index.md#organization_id) `organization_id`)The organization ID the Project is associated with. Any change made to the `organization_id` will recreate the resource.

## Attributes Reference

The `scaleway_account_project` resource exports certain attributes once the Project information is retrieved. These attributes can be referenced in other parts of your Terraform configuration.

In addition to all arguments above, the following attributes are exported:

- `id` - The unique identifier of the project (UUID format).
- `created_at` - The creation time of the Project.
- `updated_at` - The last update time of the Project.

## Import

Projects can be imported using the `id` argument, as shown below:

```bash
terraform import scaleway_account_project.project 11111111-1111-1111-1111-111111111111
```
