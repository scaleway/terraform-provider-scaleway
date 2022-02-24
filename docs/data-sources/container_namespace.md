---
page_title: "Scaleway: scaleway_container_namespace"
description: |-
Gets information about a container namespace.
---

# scaleway_container_namespace

Gets information about a container namespace.

## Example Usage

```hcl
// Get info by namespace name
data "scaleway_container_namespace" "by_name" {
  name = "my-namespace-name"
}

// Get info by namespace ID
data "scaleway_container_namespace" "my_namespace" {
  namespace_id = "11111111-1111-1111-1111-111111111111"
}
```

## Argument Reference

- `name` - (Optional) The namespace name.
  Only one of `name` and `namespace_id` should be specified.

- `namespace_id` - (Optional) The namespace id.
  Only one of `name` and `namespace_id` should be specified.
- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions) in which the namespace exists.
- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the namespace is associated with.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the Registry Namespace.
- `organization_id` - The organization ID the namespace is associated with.
- `description` - The description of the namespace.
- `environment_variables` - The environment variables of the namespace.
- `registry_endpoint` - The registry endpoint of the namespace.
- `registry_namespace_id` - The registry namespace ID of the namespace.
