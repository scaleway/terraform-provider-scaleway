---
page_title: "Scaleway: scaleway_account_project"
description: |-
Manages Scaleway Account project.
---

# scaleway_account_project

Manages organization's projects on Scaleway.

## Example Usage

```hcl
resource "scaleway_account_project" "project" {
  name = "project"
}
```

## Arguments Reference

The following arguments are supported:

- `name` - (Optional) The name of the Project.
- `description` - (Optional) The description of the Project.
- `organization_id` - (Optional. Defaults to [provider](../index.md#organization_id) `organization_id`)The organization ID the Project is associated with. Please note that any change in `organization_id` will recreate the resource.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the project (UUID format).
- `created_at` - The Project creation time.
- `updated_at` - The Project last update time.

## Import

Projects can be imported using the `id`, e.g.

```bash
$ terraform import scaleway_account_project.project 11111111-1111-1111-1111-111111111111
```
