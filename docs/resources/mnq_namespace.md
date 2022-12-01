---
page_title: "Scaleway: scaleway_mnq_namespace"
description: |-
Manages Scaleway Messaging and queuing Namespaces.
---

# scaleway_mnq_namespace

Creates and manages Scaleway Messaging and queuing Namespace.
For further information please check
our [documentation](https://pkg.go.dev/github.com/scaleway/scaleway-sdk-go@master/api/mnq/v1alpha1#pkg-index)

## Examples

### Basic

```hcl
resource "scaleway_mnq_namespace" "main" {
  name     = "test-mnq-ns"
  protocol = "nats"
  region   = "fr-par"
}
```

## Arguments Reference

The following arguments are supported:

- `name` - (Optional) The unique name of the container namespace.

- `protocol` - (Required) The protocol to apply on your namespace. Please check our
  supported [protocols](https://pkg.go.dev/github.com/scaleway/scaleway-sdk-go@master/api/mnq/v1alpha1#pkg-constants).

- `endpoint` - (Computed) The endpoint of the service matching the Namespace protocol.

- `region` - (Defaults to [provider](../index.md#region) `region`). The [region](../guides/regions_and_zones.md#regions)
  in which the namespace should be created.

- `project_id` - (Optional. Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the
  namespace is associated with.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the namespace
- `created_at` - The date and time the Namespace was created.
- `updated_at` - The date and time the Namespace was updated.

## Import

Namespaces can be imported using the `{region}/{id}`, e.g.

```bash
$ terraform import scaleway_mnq_namespace.main fr-par/11111111111111111111111111111111
```
