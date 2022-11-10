---
page_title: "Scaleway: scaleway_account_project"
description: |-
Manages Scaleway Account project.
---

# scaleway_account_project

| WARNING: This resource is in beta version. If your are in the beta group, please set the variable `SCW_ENABLE_BETA=true` in your `env` in order to use this resource. |
|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------|

Manages organization's projects on Scaleway.

## Example Usage

```hcl
resource "scaleway_account_project" "project" {
  name = "myproject"
}
```

## Arguments Reference

The following arguments are supported:

- `name` - (Optional) The name of the Project.
- `description` - (Optional) The description of the Project.
- `organization_id` - The organization ID the Project is associated with.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `created_at` - The Project creation time.
- `updated_at` - The Project last update time.

## Import

Projects can be imported using the `id`, e.g.

```bash
$ terraform import scaleway_account_project.project 11111111-1111-1111-1111-111111111111
```
