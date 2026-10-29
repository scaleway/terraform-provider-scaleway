---
subcategory: "MessageQ"
page_title: "Scaleway: scaleway_messageq_node_type"
---

# Data Source: scaleway_messageq_node_type

Gets information about an available MessageQ node type.

## Example Usage

```terraform
data "scaleway_messageq_node_type" "main" {
  name = "MESSAGEQ-SHARED-2C-8G"
}
```

## Argument Reference

- `name` - (Required) The node type name.
- `region` - (Defaults to [provider](../index.md#arguments-reference) `region`) The region in which the node type is available.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the node type in the `{region}/{name}` format.
- `description` - Description of the node type.
- `vcpus` - Number of vCPUs available.
- `memory_size_in_gb` - Amount of memory available in GB.
- `stock_status` - Stock status of the node type.
- `disabled` - Whether the node type is disabled.
- `beta` - Whether the node type is in beta.
- `instance_range` - Instance range associated with the node type offer.
- `available_volume_types` - Available storage options for the node type.
    - `type` - Volume type.
    - `description` - Volume type description.
    - `min_size_in_gb` - Minimum volume size in GB.
    - `max_size_in_gb` - Maximum volume size in GB.
    - `chunk_size_in_gb` - Volume size increment in GB.
