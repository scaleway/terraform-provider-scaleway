---
page_title: "Scaleway: scaleway_container_namespace"
description: |-
Manages Scaleway Container Namespaces.
---

# scaleway_container_namespace

Creates and manages Scaleway Container Namespace.
For more information see [the documentation](https://developers.scaleway.com/en/products/containers/api/#namespaces-cdce79).

## Examples

### Basic

```hcl
resource "scaleway_container_namespace" "main" {
  name        = "main-container-namespace"
  description = "Main container namespace"
}
```

## Arguments Reference

The following arguments are supported:

- `name` - (Required) The unique name of the container namespace.

~> **Important** Updates to `name` will recreate the namespace.

- `description` (Optional) The description of the namespace.

- `region` - (Defaults to [provider](../index.md#region) `region`). The [region](../guides/regions_and_zones.md#regions) in which the namespace should be created.

- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the namespace is associated with.

- `environment_variables` - The environment variables of the namespace.

- `secret_environment_variables` - The secret environment variables of the namespace.

- `destroy_registry` - (Defaults to false). Destroy linked container registry on deletion.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the namespace
- `organization_id` - The organization ID the namespace is associated with.
- `registry_endpoint` - The registry endpoint of the namespace.
- `registry_namespace_id` - The registry namespace ID of the namespace.


## Import

Namespaces can be imported using the `{region}/{id}`, e.g.

```bash
$ terraform import scaleway_container_namespace.main fr-par/11111111-1111-1111-1111-111111111111
```
