---
layout: "scaleway"
page_title: "Scaleway: scaleway_container_registry"
description: |-
  Manages Scaleway Container Registries.
---

# scaleway_container_registry

Creates and manages Scaleway Container Registry. For more information see [the documentation](https://developers.scaleway.com/en/products/registry/api/).

## Examples

### Basic

```hcl
resource "scaleway_container_registry" "main" {
    region = "fr-par"
    name = "main_cr"
    description = "Main container registry"
    is_public = false
}
```

## Arguments Reference

The following arguments are supported:

- `name` - (Required) The unique name of the container registry namespace.

~> **Important** Updates to `name` will recreate the namespace.

- `region` - (Defaults to [provider](../index.html#region) `region`). The [region](../guides/regions_and_zones.html#regions) in which the container registry namespace should be created.

- `organization_id` - (Defaults to [provider](../index.html#organization_id) `organization_id`) The ID of the organization the registry is associated with.

- `description` (Optional) The description of the container registry namespace.

- `is_public` (Defaults to `false`) Whether or not the registry should be visible to the public.

## Attibutes Reference

In addition to all arguments above, the following attibutes are exported:

- `id` - The ID of the namespace
- `endpoint` - Endpoint reachable by docker.

## Import

Container Registry Namespace can be imported using the `{region}/{id}`, eg.

```bash
$ terraform import scaleway_container_registy.cr01 fr-par/11111111-1111-1111-1111-111111111111
```