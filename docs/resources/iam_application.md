---
page_title: "Scaleway: scaleway_iam_application"
description: |-
Manages Scaleway IAM Applications.
---

# scaleway_iam_application

| WARNING: This resource is in beta version. If your are in the beta group, please set the variable `SCW_ENABLE_BETA=true` in your `env` in order to use this resource. |
|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------|


Creates and manages Scaleway IAM Applications. For more information, see [the documentation](https://developers.scaleway.com/en/products/iam/api/v1alpha1/#applications-83ce5e).

## Example Usage

```hcl
resource "scaleway_iam_application" "main" {
  name = "My application"
  description = "a description"
}
```

## Arguments Reference

The following arguments are supported:

- `name` - .The name of the iam application.
- `description` - The description of the iam application.
- `organization_id` - (Defaults to [provider](../index.md#organization_d) `organization_id`) The ID of the organization the application is associated with.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `created_at` - The date and time of the creation of the application.
- `updated_at` - The date and time of the last update of the application.
- `editable` - Whether the application is editable.

## Import

Applications can be imported using the `{id}`, e.g.

```bash
$ terraform import scaleway_iam_application.main 11111111-1111-1111-1111-111111111111
```
